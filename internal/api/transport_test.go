package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryTransport_RetryOn429(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	transport := NewRetryTransport(srv.Client().Transport)
	transport.MaxRetries = 3

	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestRetryTransport_NoRetryOn400(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(400)
	}))
	defer srv.Close()

	transport := NewRetryTransport(srv.Client().Transport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if attempts.Load() != 1 {
		t.Fatalf("expected 1 attempt (no retry), got %d", attempts.Load())
	}
}

func TestRetryTransport_RetryAfterHeader(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	transport := NewRetryTransport(srv.Client().Transport)

	start := time.Now()
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := transport.RoundTrip(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	// Retry-After=1 means at least 1s backoff (same as default first backoff).
	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected at least ~1s delay for Retry-After, got %v", elapsed)
	}
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(503)
	}))
	defer srv.Close()

	transport := NewRetryTransport(srv.Client().Transport)
	transport.MaxRetries = 1

	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Fatalf("expected 503 after exhausted retries, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(429)
	}))
	defer srv.Close()

	transport := NewRetryTransport(srv.Client().Transport)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Fatal("expected context error during long backoff")
	}
}

func TestParseRetryAfterHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"empty", "", 0},
		{"integer", "5", 5 * time.Second},
		{"zero", "0", 0},
		{"garbage", "not-a-number", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfterHeader(tt.header)
			if got != tt.want {
				t.Errorf("parseRetryAfterHeader(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}
