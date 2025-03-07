package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

// Mode represents the application runtime mode.
type Mode string

const (
	Prod Mode = "prod"
	Dev  Mode = "dev"
)

// Config holds the application configuration.
type Config struct {
	App  App
	Mail Mail
	Data Data
}

type App struct {
	Version       string     `env:"APP_VERSION" env-default:"0.0.1"`
	LogLevel      slog.Level `env:"APP_LOG_LEVEL" env-default:"info"`
	Mode          Mode       `env:"APP_MODE" env-default:"prod"`
	MaxGoroutines int        `env:"APP_MAX_GOROUTINES" env-default:"5"`
}

type Mail struct {
	From         string         `env:"MAIL_FROM"`
	Host         string         `env:"MAIL_HOST"`
	Password     string         `env:"MAIL_PASSWORD"`
	Port         int            `env:"MAIL_PORT"`
	To           []string       `env:"MAIL_TO"`
	MailStores   map[int]string `env:"MAIL_STORES"`
	Subject      string         `env:"MAIL_SUBJECT"`
	TemplateName string         `env:"MAIL_TEMPLATE_NAME"`
}

type Data struct {
	Url               url.URL           `env:"DATA_URL"`
	ApiKey            string            `env:"DATA_API_KEY"`
	IgnoredGroups     []string          `env:"DATA_IGNORED_GROUPS"`    // DATA_IGNORED_GROUPS='group01,group02,group with spaces'
	Companies         map[string]string `env:"DATA_COMPANIES"`         // DATA_COMPANIES='key01:value01,key with space:value with space'
	AllowedCompanies  []string          `env:"DATA_ALLOWED_COMPANIES"` // DATA_DATA_ALLOWED_COMPANIES='company01,company with spaces'
	MaxOffline        time.Duration     `env:"DATA_MAX_OFFLINE"`       // DATA_MAX_OFFLINE=48h
	StoreTestNumber   int               `env:"DATA_STORE_TEST_NUMBER"`
	StoreNumberPrefix string            `env:"DATA_STORE_NUMBER_PREFIX"`
	CompanyNamePrefix string            `env:"DATA_COMPANY_NAME_PREFIX"`
	IgnoredTags       []string          `env:"DATA_IGNORED_TAGS"`
}

// Must load the configuration and panics if it fails.
// Use this when configuration is required for the application to start.
func Must() Config {
	var config Config

	if err := cleanenv.ReadEnv(&config); err != nil {
		panic(fmt.Sprintf("Error processing environment variables: %v", err))
	}

	return config
}
