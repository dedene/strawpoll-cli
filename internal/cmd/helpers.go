package cmd

import (
	"fmt"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/auth"
)

// newClientFromAuth creates an API client using the stored API key.
func newClientFromAuth() (*api.Client, error) {
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	return api.NewClient(apiKey), nil
}
