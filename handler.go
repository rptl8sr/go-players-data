package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go-players-data/internal/cluster"
	"go-players-data/internal/config"
	"go-players-data/internal/fetcher"
	"go-players-data/internal/filter"
	"go-players-data/internal/logger"
	"go-players-data/internal/mailer"
	"go-players-data/internal/model"
	"go-players-data/internal/player"
	"go-players-data/internal/templateloader"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

func Handler() (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	defer func() { logger.Info("main.Handler: Time spent", "time", time.Since(start).String()) }()

	cfg := config.Must()
	logger.Init(cfg.App.LogLevel)
	logger.Info("main.Handler: Starting")

	if cfg.App.Mode == config.Dev {
		logger.Debug("main.Handler: Config", "cfg", cfg)
	}

	dataFetcher := fetcher.New(http.DefaultClient, cfg.Data.Url, cfg.Data.ApiKey)
	playerParser := player.New(cfg.Data)
	filterCriteria := filter.New(cfg.Data.IgnoredGroups, cfg.Data.AllowedCompanies, cfg.Data.MaxOffline)
	clusterProcessor := cluster.New()
	templateLoader, err := templateloader.New()
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}
	mailProcessor, err := mailer.New(cfg.Mail, templateLoader)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	body, err := dataFetcher.Data(ctx)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	allPlayers, err := playerParser.Players(body)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	players, err := filterCriteria.Filter(allPlayers)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	clusters := clusterProcessor.ByStoreNumber(players)

	mailByCluster(
		mailProcessor,
		clusters,
		cfg.App.MaxGoroutines,
	)

	logger.Debug("main.Handler", "offline_players", len(players), "all_players", len(allPlayers))

	return &Response{
		StatusCode: 200,
		Body:       "Successful response",
	}, nil
}

func mailByCluster(m mailer.Mailer, clusters map[int][]*model.Player, maxGoroutines int) {
	start := time.Now()
	defer func() { logger.Debug("main.mailByCluster: Time spent", "time", time.Since(start).String()) }()

	sem := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	for storeNumber, clusterPlayers := range clusters {
		sem <- struct{}{}
		wg.Add(1)

		go func(sn int, players []*model.Player) {
			defer func() {
				<-sem
				wg.Done()
			}()

			if err := m.Send(storeNumber, players); err != nil {
				logger.Error("main.Handler: Failed to send mail",
					"err", err,
					"cluster", storeNumber,
					"players", len(clusterPlayers),
				)
			}
		}(storeNumber, clusterPlayers)
	}

	wg.Wait()
}
