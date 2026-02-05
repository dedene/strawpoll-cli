package api

import (
	"context"
	"fmt"
)

// CreatePoll creates a new poll via POST /polls.
// Sets type to "multiple_choice" if not already set.
func (c *Client) CreatePoll(ctx context.Context, req *CreatePollRequest) (*Poll, error) {
	if req.Type == "" {
		req.Type = PollTypeMultipleChoice
	}

	var poll Poll
	if err := c.Post(ctx, "/polls", req, &poll); err != nil {
		return nil, fmt.Errorf("create poll: %w", err)
	}

	return &poll, nil
}

// GetPoll retrieves a poll by ID via GET /polls/{id}.
func (c *Client) GetPoll(ctx context.Context, id string) (*Poll, error) {
	var poll Poll
	if err := c.Get(ctx, "/polls/"+id, &poll); err != nil {
		return nil, fmt.Errorf("get poll: %w", err)
	}

	return &poll, nil
}

// GetPollResults retrieves poll results via GET /polls/{id}/results.
func (c *Client) GetPollResults(ctx context.Context, id string) (*PollResults, error) {
	var results PollResults
	if err := c.Get(ctx, "/polls/"+id+"/results", &results); err != nil {
		return nil, fmt.Errorf("get poll results: %w", err)
	}

	return &results, nil
}

// DeletePoll deletes a poll via DELETE /polls/{id}.
func (c *Client) DeletePoll(ctx context.Context, id string) error {
	if err := c.Delete(ctx, "/polls/"+id); err != nil {
		return fmt.Errorf("delete poll: %w", err)
	}

	return nil
}

// UpdatePoll updates a poll via PUT /polls/{id}.
func (c *Client) UpdatePoll(ctx context.Context, id string, req *UpdatePollRequest) (*Poll, error) {
	var poll Poll
	if err := c.Put(ctx, "/polls/"+id, req, &poll); err != nil {
		return nil, fmt.Errorf("update poll: %w", err)
	}

	return &poll, nil
}

// ResetPollResults resets poll results via DELETE /polls/{id}/results.
func (c *Client) ResetPollResults(ctx context.Context, id string) error {
	if err := c.Delete(ctx, "/polls/"+id+"/results"); err != nil {
		return fmt.Errorf("reset poll results: %w", err)
	}

	return nil
}

// ListMyPolls lists the user's polls via GET /users/@me/polls.
// pollType must be "created" or "participated".
func (c *Client) ListMyPolls(ctx context.Context, pollType string, page, limit int) (*PollListResponse, error) {
	path := fmt.Sprintf("/users/@me/polls?type=%s&page=%d&limit=%d", pollType, page, limit)

	var resp PollListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list polls: %w", err)
	}

	return &resp, nil
}
