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

type Request struct {
	APIKey string `json:"report_api_key"`
}

type fetcher struct {
	url    url.URL
	token  string
	client *http.Client
}

type Fetcher interface {
	Data(ctx context.Context) ([]byte, error)
}

func New(c *http.Client, u url.URL, token string) Fetcher {
	return &fetcher{
		url:    u,
		token:  token,
		client: c,
	}
}

// Data sends a POST request to the specified URL with a JSON payload
// containing the provided API token to receive players' data.
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

type HTTPError struct {
	Code int
}

func (e *HTTPError) Error() string {
	return http.StatusText(e.Code)
}
