package main

import (
	"context"
	"encoding/json"
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

// TimerEvent represents the structure of an event from a Yandex Cloud timer trigger.
type TimerEvent struct {
	ID          string `json:"id"`
	TriggerType string `json:"trigger_type"`
	TriggeredAt string `json:"triggered_at"`
}

// HTTPEvent represents the structure of an event from a Yandex Cloud HTTP trigger.
type HTTPEvent struct {
	HTTPMethod      string            `json:"http_method"`
	Path            string            `json:"path"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"is_base64_encoded"`
}

// Response defines the response format for the Yandex Cloud Function.
// Used for HTTP triggers; ignored for timer triggers.
type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

// Handler is the entry point for the Yandex Cloud Function.
// Processes events from timer or HTTP triggers, fetches player data,
// filters it, and sends notifications by clusters.
func Handler(ctx context.Context, event interface{}) (*Response, error) {
	start := time.Now()
	defer func() { logger.Info("main.Handler: Time spent", "time", time.Since(start).String()) }()

	cfg := config.Must()
	triggerType := detectTriggerType(event)
	logger.Init(cfg.App.LogLevel)
	logger.Info("main.Handler: Starting", "trigger_type", triggerType)

	if cfg.App.Mode == config.Dev {
		logger.Debug("main.Handler: Config", "cfg", cfg)
	}

	// Initialize dependencies for data processing
	dataFetcher := fetcher.New(http.DefaultClient, cfg.Data.Url, cfg.Data.ApiKey)
	playerParser := player.New(cfg.Data)
	filterCriteria := filter.New(cfg.Data.IgnoredGroups, cfg.Data.AllowedCompanies, cfg.Data.IgnoredTags, cfg.Data.MaxOffline)
	clusterProcessor := cluster.New()

	// Load email templates
	templateLoader, err := templateloader.New()
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}
	// Initialize mail processor
	mailProcessor, err := mailer.New(cfg.Mail, templateLoader)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	// Fetch player data from an external source
	body, err := dataFetcher.Data(ctx)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	// Parse all players from the fetched data
	allPlayers, err := playerParser.Players(body)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	// Filter players based on specified criteria
	players, err := filterCriteria.Filter(allPlayers)
	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, err
	}

	// Group players by store number
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

// mailByCluster sends notifications for player clusters in parallel goroutines.
// Uses semaphore to limit the number of concurrent tasks.
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

			if err := m.Send(sn, players); err != nil {
				logger.Error("main.Handler: Failed to send mail",
					"err", err,
					"cluster", sn,
					"players", len(players),
				)
			}
		}(storeNumber, clusterPlayers)
	}

	wg.Wait()
}

// detectTriggerType determines the type of trigger that invoked the function (timer or HTTP).
// Returns "timer", "http", or "unknown" if the event type is not recognized.
func detectTriggerType(event interface{}) string {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return "unknown"
	}

	var timerEvent TimerEvent
	if json.Unmarshal(eventBytes, &timerEvent) == nil && timerEvent.TriggerType == "TIMER" {
		return "timer"
	}

	var httpEvent HTTPEvent
	if json.Unmarshal(eventBytes, &httpEvent) == nil && httpEvent.HTTPMethod != "" {
		return "http"
	}

	return "unknown"
}
