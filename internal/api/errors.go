package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// APIError represents a generic API error response.
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Message)
}

// AuthError represents a 401/403 authentication error.
type AuthError struct {
	APIError
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

// RateLimitError represents a 429 rate limit error.
type RateLimitError struct {
	APIError
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited: retry after %ds", int(e.RetryAfter.Seconds()))
}

// ValidationError represents a field validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s %s", e.Field, e.Message)
}

// apiErrorResponse mirrors the StrawPoll JSON error format.
type apiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// NewAPIError parses a JSON error response body and returns the appropriate typed error.
func NewAPIError(statusCode int, body []byte) error {
	var resp apiErrorResponse
	msg := "unknown error"
	if err := json.Unmarshal(body, &resp); err == nil && resp.Error.Message != "" {
		msg = resp.Error.Message
	}

	base := APIError{
		StatusCode: statusCode,
		Message:    msg,
	}

	switch statusCode {
	case 401, 403:
		return &AuthError{APIError: base}
	case 429:
		return &RateLimitError{APIError: base, RetryAfter: parseRetryAfter(body)}
	default:
		return &base
	}
}

// parseRetryAfter attempts to extract retry-after seconds from the response body.
func parseRetryAfter(body []byte) time.Duration {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0
	}

	if v, ok := raw["retry_after"]; ok {
		switch val := v.(type) {
		case float64:
			return time.Duration(val) * time.Second
		case string:
			if secs, err := strconv.Atoi(val); err == nil {
				return time.Duration(secs) * time.Second
			}
		}
	}
	return 0
}
