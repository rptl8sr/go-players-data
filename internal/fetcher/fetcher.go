package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"go-players-data/internal/logger"
)

// Request represents the payload for requests that include an API key as a JSON field.
type Request struct {
	APIKey string `json:"report_api_key"`
}

// fetcher is a concrete implementation that fetches data from a URL using an HTTP client and an API token.
// it includes the endpoint URL, authorization token, and a pointer to the HTTP client for request execution.
type fetcher struct {
	url    url.URL
	token  string
	client *http.Client
}

// Fetcher is an interface for retrieving data, requiring a method to get it with context handling for cancellations.
type Fetcher interface {
	Data(ctx context.Context) ([]byte, error)
}

// New creates a new Fetcher instance with the provided HTTP client, URL, and API key.
func New(c *http.Client, u url.URL, token string) Fetcher {
	return &fetcher{
		url:    u,
		token:  token,
		client: c,
	}
}

// Data fetches data from the configured URL with the API key in the Authorization header.
// Respects the provided context for cancellation and timeouts.
func (f *fetcher) Data(ctx context.Context) ([]byte, error) {
	start := time.Now()
	defer func() { logger.Debug("fetcher.FetchData: Time spent", "time", time.Since(start).String()) }()

	data, err := json.Marshal(Request{
		APIKey: f.token,
	})
	if err != nil {
		logger.Error("fetcher.FetchData: Error marshaling request", "err", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.url.String(), bytes.NewBuffer(data))
	if err != nil {
		logger.Error("fetcher.FetchData: Error creating request", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		logger.Error("fetcher.FetchData: Error sending request", "err", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		logger.Error("fetcher.FetchData: Invalid status code", "statusCode", resp.StatusCode)
		return nil, &HTTPError{Code: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("fetcher.FetchData: Error reading response body", "err", err)
		return nil, err
	}

	return body, nil
}

// HTTPError represents an error response from an HTTP request with a specific status code.
type HTTPError struct {
	Code int
}

// Error returns the text representation of the HTTP status code associated with the HTTPError.
func (e *HTTPError) Error() string {
	return http.StatusText(e.Code)
}
