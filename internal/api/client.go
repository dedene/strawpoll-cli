package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.strawpoll.com/v3"
	defaultTimeout = 30 * time.Second
	defaultRate    = 10
	rateInterval   = time.Second
)

// Client is the StrawPoll API client.
type Client struct {
	httpClient  *http.Client
	rateLimiter *RateLimiter
	apiKey      string
	baseURL     string
}

// NewClient creates a Client with retry transport, rate limiter, and auth.
func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: NewRetryTransport(http.DefaultTransport),
			Timeout:   defaultTimeout,
		},
		rateLimiter: NewRateLimiter(defaultRate, rateInterval),
		apiKey:      apiKey,
		baseURL:     defaultBaseURL,
	}
}

// Close releases resources held by the client.
func (c *Client) Close() {
	c.rateLimiter.Close()
}

// do executes an API request with rate limiting, auth, and error handling.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
