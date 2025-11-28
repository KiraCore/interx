package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is an HTTP client for Interx API testing
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Interx API client
func NewClient(cfg Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Response represents an API response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// JSON unmarshals the response body into the given interface
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// IsSuccess returns true if status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Get performs a GET request
func (c *Client) Get(path string, queryParams map[string]string) (*Response, error) {
	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		params := url.Values{}
		for k, v := range queryParams {
			params.Add(k, v)
		}
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req)
}

// Post performs a POST request
func (c *Client) Post(path string, body interface{}) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req)
}

func (c *Client) do(req *http.Request) (*Response, error) {
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	_ = time.Since(start) // Can be used for performance metrics

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// HealthCheck performs a basic health check
func (c *Client) HealthCheck() error {
	resp, err := c.Get("/api/status", nil)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}
	return nil
}
