package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testServer(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := NewClient("test-api-key")
	c.baseURL = srv.URL

	return c
}

func TestCreatePoll(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	var gotBody CreatePollRequest
	var gotAPIKey string

	c := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/polls" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		gotAPIKey = r.Header.Get("X-API-Key")

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		resp := Poll{
			ID:    "abc123",
			Title: "Favorite Color",
			Type:  PollTypeMultipleChoice,
			PollOptions: []*PollOption{
				{ID: "opt1", Value: "Red"},
				{ID: "opt2", Value: "Blue"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	req := &CreatePollRequest{
		Title: "Favorite Color",
		PollOptions: []*PollOption{
			{Value: "Red"},
			{Value: "Blue"},
		},
	}

	poll, err := c.CreatePoll(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePoll: %v", err)
	}

	if gotAPIKey != "test-api-key" {
		t.Errorf("X-API-Key = %q, want %q", gotAPIKey, "test-api-key")
	}

	if gotBody.Type != PollTypeMultipleChoice {
		t.Errorf("request type = %q, want %q", gotBody.Type, PollTypeMultipleChoice)
	}

	if poll.ID != "abc123" {
		t.Errorf("poll.ID = %q, want %q", poll.ID, "abc123")
	}

	if poll.Title != "Favorite Color" {
		t.Errorf("poll.Title = %q, want %q", poll.Title, "Favorite Color")
	}

	if len(poll.PollOptions) != 2 {
		t.Errorf("poll.PollOptions = %d, want 2", len(poll.PollOptions))
	}
}

func TestGetPoll(t *testing.T) {
	c := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/polls/xyz789" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		resp := Poll{
			ID:    "xyz789",
			Title: "Best Language",
			Type:  PollTypeMultipleChoice,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	poll, err := c.GetPoll(context.Background(), "xyz789")
	if err != nil {
		t.Fatalf("GetPoll: %v", err)
	}

	if poll.ID != "xyz789" {
		t.Errorf("poll.ID = %q, want %q", poll.ID, "xyz789")
	}

	if poll.Title != "Best Language" {
		t.Errorf("poll.Title = %q, want %q", poll.Title, "Best Language")
	}
}

func TestGetPollResults(t *testing.T) {
	c := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/polls/abc123/results" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		resp := PollResults{
			ID:               "abc123",
			VoteCount:        42,
			ParticipantCount: 30,
			PollOptions: []*PollOption{
				{ID: "opt1", Value: "Red", VoteCount: 25},
				{ID: "opt2", Value: "Blue", VoteCount: 17},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	results, err := c.GetPollResults(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("GetPollResults: %v", err)
	}

	if results.VoteCount != 42 {
		t.Errorf("VoteCount = %d, want 42", results.VoteCount)
	}

	if results.ParticipantCount != 30 {
		t.Errorf("ParticipantCount = %d, want 30", results.ParticipantCount)
	}

	if len(results.PollOptions) != 2 {
		t.Errorf("PollOptions = %d, want 2", len(results.PollOptions))
	}
}

func TestDeletePoll(t *testing.T) {
	c := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/polls/del456" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeletePoll(context.Background(), "del456")
	if err != nil {
		t.Fatalf("DeletePoll: %v", err)
	}
}

func TestCreatePollAuthError(t *testing.T) {
	c := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Invalid API key",
				"code":    401,
			},
		})
	}))

	_, err := c.CreatePoll(context.Background(), &CreatePollRequest{
		Title:       "Test",
		PollOptions: []*PollOption{{Value: "A"}},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var authErr *AuthError
	if !errorAs(err, &authErr) {
		t.Errorf("expected AuthError, got %T: %v", err, err)
	}
}

// errorAs is a helper that wraps errors.As for unwrapping fmt.Errorf %w chains.
func errorAs[T any](err error, target *T) bool {
	for err != nil {
		if e, ok := any(err).(T); ok { //nolint:errorlint // intentional type assertion
			*target = e

			return true
		}

		if e, ok := any(err).(interface{ Unwrap() error }); ok {
			err = e.Unwrap()
		} else {
			return false
		}
	}

	return false
}
