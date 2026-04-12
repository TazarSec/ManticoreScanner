package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// AuthError indicates a 401 response.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string { return fmt.Sprintf("authentication failed: %s", e.Message) }

// RateLimitError indicates a 429 response.
type RateLimitError struct {
	Message      string
	RetryAfterSec int
}

func (e *RateLimitError) Error() string { return fmt.Sprintf("rate limited: %s", e.Message) }

// ValidationError indicates a 400 response.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation error: %s", e.Message) }

// ServerError indicates an unexpected server error.
type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error (%d): %s", e.StatusCode, e.Message)
}

// Client communicates with the AegisEngine scan API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// ScanBatch submits a batch of packages for scanning.
// Returns the batch response and the HTTP status code (200 = all done, 202 = some pending).
func (c *Client) ScanBatch(ctx context.Context, items []ScanRequestItem) (*BatchResponse, int, error) {
	reqBody := ScanRequest{Packages: items}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/scan", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("reading response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		var batch BatchResponse
		if err := json.Unmarshal(respBody, &batch); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("decoding response: %w", err)
		}
		return &batch, resp.StatusCode, nil

	case http.StatusBadRequest:
		var apiErr APIError
		_ = json.Unmarshal(respBody, &apiErr)
		return nil, resp.StatusCode, &ValidationError{Message: apiErr.Error}

	case http.StatusUnauthorized:
		var apiErr APIError
		_ = json.Unmarshal(respBody, &apiErr)
		return nil, resp.StatusCode, &AuthError{Message: apiErr.Error}

	case http.StatusTooManyRequests:
		var apiErr APIError
		_ = json.Unmarshal(respBody, &apiErr)
		retryAfter := 0
		if val := resp.Header.Get("Retry-After"); val != "" {
			retryAfter, _ = strconv.Atoi(val)
		}
		return nil, resp.StatusCode, &RateLimitError{
			Message:      apiErr.Error,
			RetryAfterSec: retryAfter,
		}

	default:
		var apiErr APIError
		_ = json.Unmarshal(respBody, &apiErr)
		return nil, resp.StatusCode, &ServerError{
			StatusCode: resp.StatusCode,
			Message:    apiErr.Error,
		}
	}
}
