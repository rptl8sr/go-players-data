package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"go-players-data/internal/config"
	"go-players-data/internal/logger"
	"go-players-data/internal/model"
	"go-players-data/internal/templateloader"
)

// mailer is a struct used for managing email configurations and rendering email templates.
type mailer struct {
	config config.Mail
	tmpl   *template.Template
}

// mailData represents the structure for email-related data including sender, recipients, subject, store details, and players.
type mailData struct {
	From        string
	To          []string
	Subject     string
	StoreNumber int
	StoreID     string
	Players     []*model.Player
}

// Mailer defines an interface for sending email notifications to players grouped by store number.
type Mailer interface {
	Send(storeNumber int, players []*model.Player) error
}

// New initializes a Mailer instance with the given configuration and template loader.
// It loads the mail template using the specified template name and custom template functions.
// Returns a configured Mailer instance or an error if template initialization fails.
func New(cfg config.Mail, loader *templateloader.Loader) (Mailer, error) {
	tmpl, err := loader.Load(
		cfg.TemplateName,
		template.FuncMap{
			"join": strings.Join,
			"base64enc": func(s string) string {
				return base64.StdEncoding.EncodeToString([]byte(s))
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("mailer.New: mail template initialization failed: %w", err)
	}

	return &mailer{
		config: cfg,
		tmpl:   tmpl,
	}, nil
}

// Send constructs and sends an email using the specified store number and player details. Returns an error if it fails.
func (m *mailer) Send(storeNumber int, players []*model.Player) error {
	start := time.Now()
	defer func() { logger.Debug("mailer.Send: Time spent", "time", time.Since(start).String()) }()

	body, err := m.body(storeNumber, players)
	if err != nil {
		return fmt.Errorf("mailer.Send: failed to build mail body: %w", err)
	}

	if err = m.send(body); err != nil {
		return fmt.Errorf("mailer.Send: failed to send mail: %w", err)
	}

	return nil
}

// send sends an email with the specified body using the configured SMTP server and authentication.
// returns an error on failure.
func (m *mailer) send(body string) error {
	auth := smtp.PlainAuth("", m.config.From, m.config.Password, m.config.Host)
	return smtp.SendMail(
		fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		auth,
		m.config.From,
		m.config.To,
		[]byte(body),
	)
}

// body generates the email body using the provided store number and player details, returning it as a string or an error.
func (m *mailer) body(storeNumber int, players []*model.Player) (string, error) {
	var storeID string

	if m.config.MailStores[storeNumber] != "" {
		storeID = m.config.MailStores[storeNumber]
	} else {
		storeID = fmt.Sprintf("%d", storeNumber)
	}

	var buf bytes.Buffer

	data := &mailData{
		From:        m.config.From,
		To:          m.config.To,
		Subject:     m.config.Subject,
		StoreNumber: storeNumber,
		StoreID:     storeID,
		Players:     players,
	}

	if err := m.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("mailer.body: failed to execute template: %w", err)
	}

	return buf.String(), nil
}
