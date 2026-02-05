package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient creates a Client pointing at a test server.
func newTestClient(srv *httptest.Server) *Client {
	return &Client{
		httpClient:  srv.Client(),
		rateLimiter: NewRateLimiter(100, time.Second),
		apiKey:      "test-api-key",
		baseURL:     srv.URL,
	}
}

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("missing or wrong API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "abc123"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer c.Close()

	var out map[string]string
	err := c.Get(context.Background(), "/polls/abc123", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["id"] != "abc123" {
		t.Fatalf("expected id=abc123, got %s", out["id"])
	}
}

func TestClient_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "Test Poll" {
			t.Errorf("expected title=Test Poll, got %s", body["title"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "new123"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer c.Close()

	var out map[string]string
	err := c.Post(context.Background(), "/polls", map[string]string{"title": "Test Poll"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["id"] != "new123" {
		t.Fatalf("expected id=new123, got %s", out["id"])
	}
}

func TestClient_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer c.Close()

	err := c.Delete(context.Background(), "/polls/abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Invalid API key",
				"code":    401,
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer c.Close()

	err := c.Get(context.Background(), "/polls", nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T: %v", err, err)
	}
}

func TestClient_NoAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "" {
			t.Error("expected no API key header for empty key")
		}
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.apiKey = ""
	defer c.Close()

	err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Put(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer c.Close()

	var out map[string]string
	err := c.Put(context.Background(), "/polls/abc123", map[string]string{"title": "Updated"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["updated"] != "true" {
		t.Fatalf("expected updated=true, got %s", out["updated"])
	}
}
