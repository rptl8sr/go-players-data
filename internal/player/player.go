package player

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-players-data/internal/config"
	"go-players-data/internal/logger"
	"go-players-data/internal/model"
)

// ErrParseID is returned when an error occurs while parsing or converting the ID field from input data.
// ErrParseTZ is returned when an error occurs while parsing or converting the time zone from input data.
// ErrParseLastOnline is returned when an error occurs while parsing the "last online" timestamp from input data.
var (
	ErrParseID         = errors.New("error parsing id")
	ErrParseTZ         = errors.New("error parsing time zone") // ErrParseLastOnline is returned when an error occurs while parsing the "last online" timestamp from input data.
	ErrParseLastOnline = errors.New("error parsing last online")
)

// parser is a struct that provides functionality to parse and transform data into structured and validated formats.
type parser struct {
	storeTestNumber   int
	storeNumberPrefix string
	companyNamePrefix string
	companies         map[string]string
}

// Parser is an interface for parsing raw byte data into structured player objects.
type Parser interface {
	Players(body []byte) ([]*model.Player, error)
}

// New initializes and returns a new Parser instance configured with the provided configuration data.
// It ensures that the Companies map is not nil, creating a new map if necessary.
func New(cfg config.Data) Parser {
	if cfg.Companies == nil {
		cfg.Companies = make(map[string]string)
	}
	return &parser{
		storeTestNumber:   cfg.StoreTestNumber,
		storeNumberPrefix: cfg.StoreNumberPrefix,
		companyNamePrefix: cfg.CompanyNamePrefix,
		companies:         cfg.Companies,
	}
}

// Players parse raw player data from the provided byte slice
// using the specified configuration and return a slice of players.
func (p *parser) Players(body []byte) ([]*model.Player, error) {
	start := time.Now()
	defer func() { logger.Debug("parser.Players: Time spent", "time", time.Since(start).String()) }()

	rawPlayers, err := p.parseRaw(body)
	if err != nil {
		return nil, err
	}

	players, err := p.rawToPlayers(rawPlayers)
	if err != nil {
		return nil, err
	}

	return players, nil
}

// parseRaw parses raw JSON byte data into a slice of PlayerReceive objects
// and returns it or an error if unmarshalling fails.
func (p *parser) parseRaw(body []byte) ([]*model.PlayerReceive, error) {
	var rawPlayers []*model.PlayerReceive
	if err := json.Unmarshal(body, &rawPlayers); err != nil {
		logger.Error("parser.ParseRaw: Error unmarshalling raw players", "err", err)
		return nil, err
	}

	return rawPlayers, nil
}

// rawToPlayers converts a slice of raw player data (PlayerReceive)
// into a slice of validated and structured Players objects.
// It initializes each player using the provided configuration and skips entries with errors during initialization.
// Returns the resulting slice of Players objects and an error if critical processing issues occur.
func (p *parser) rawToPlayers(rawPlayers []*model.PlayerReceive) ([]*model.Player, error) {
	players := make([]*model.Player, 0, len(rawPlayers))

	for _, raw := range rawPlayers {
		player, err := p.initPlayer(raw)
		if err != nil {
			logger.Error("parser.RawToPlayer: Error initializing player", "err", err)
			continue
		}
		players = append(players, player)
	}
	return players, nil
}

// initPlayer initializes a Players object from a PlayerReceive structure
// and configuration, performing the necessary validations.
// Converts and parses fields like IDs, time zones, tags, and timestamps. Returns errors for invalid input data.
func (p *parser) initPlayer(raw *model.PlayerReceive) (*model.Player, error) {
	var id int
	var err error

	if raw.ID != "" {
		id, err = strconv.Atoi(raw.ID)
		if err != nil {
			logger.Error("parser.RawToPlayer: Error converting id to int", "err", err, "id", raw.ID)
			return nil, ErrParseID
		}
	}

	tz, err := strconv.Atoi(raw.TimeZoneDiff)
	if err != nil {
		logger.Error("parser.RawToPlayer: Error converting time zone diff to int", "err", err, "tz", raw.TimeZoneDiff)
		return nil, ErrParseTZ
	}

	lastOnline, err := time.Parse(time.DateTime, raw.LastOnline)
	if err != nil {
		logger.Error("parser.RawToPlayer: Error parsing last online", "err", err)
		return nil, ErrParseLastOnline
	}

	var tags []string
	if raw.Tags != "" {
		tags = strings.Split(raw.Tags, ",")
	}

	player := &model.Player{
		Number:       raw.Number,
		ID:           id,
		GroupName:    raw.GroupName,
		PlayerName:   raw.PlayerName,
		Tags:         tags,
		ScheduleName: raw.ScheduleName,
		TimeZoneDiff: tz,
		LastOnline:   lastOnline,
		Serial:       raw.Serial,
		MAC:          p.normalizeMAC(raw.MAC),
		IP:           raw.IP,
		Type:         raw.Type,
		Model:        raw.Model,
		Version:      raw.Version,
		StoreNumber:  0,
		CompanyName:  "",
	}

	p.parseTags(player)

	return player, nil
}

// parseTags processes the tags of a Players object to extract store numbers and company names based on defined prefixes.
// Updates the Players' store number and company name fields, using configuration data for validation and mapping.
func (p *parser) parseTags(player *model.Player) {
	for _, tag := range player.Tags {
		switch {
		case strings.HasPrefix(tag, p.storeNumberPrefix):
			numberTag := strings.TrimPrefix(tag, p.storeNumberPrefix)
			if numberTag == "" {
				logger.Debug("parser.parseTags: Empty store number tag", "player", player)
				continue
			}

			n, err := strconv.Atoi(numberTag)
			if err != nil {
				logger.Error("parser.parseTags: Error converting number tag to int", "err", err, "numberTag", numberTag, "player", player)
				continue
			}

			if n == p.storeTestNumber {
				continue
			}

			player.StoreNumber = n
		case strings.HasPrefix(tag, p.companyNamePrefix):
			companyNameTag := strings.TrimPrefix(tag, p.companyNamePrefix)
			if companyNameTag == "" {
				logger.Warn("parser.parseTags: Empty company name tag", "player", player)
				continue
			}

			v, ok := p.companies[companyNameTag]
			if !ok {
				logger.Warn("parser.parseTags: Unknown company name", "company_name", companyNameTag, "player", player)
				player.CompanyName = companyNameTag
			} else {
				player.CompanyName = v
			}
		default:
			continue
		}
	}
}

// normalizeMAC takes a raw MAC address string, removes invalid characters,
// converts to lowercase, and formats as XX:XX:XX:XX:XX:XX.
// Returns an empty string if the input is invalid or does not produce a 12-character string.
// Log a warning for invalid inputs.
func (p *parser) normalizeMAC(macRaw string) string {
	if macRaw == "" {
		return ""
	}

	mac := strings.Map(func(r rune) rune {
		if '0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F' {
			return r
		}
		return -1
	}, macRaw)

	mac = strings.ToLower(mac)

	if len(mac) != 12 {
		logger.Warn("parser.normalizeMAC: Invalid MAC address", "mac", mac)
		return ""
	}

	var builder strings.Builder
	for i := 0; i < len(mac); i += 2 {
		if i != 0 {
			builder.WriteByte(':')
		}
		builder.WriteString(mac[i : i+2])
	}
	return strings.ToUpper(builder.String())
}
