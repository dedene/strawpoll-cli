package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// RankingUpdateCmd updates an existing ranking poll.
type RankingUpdateCmd struct {
	ID           string   `arg:"" required:"" help:"Poll ID or URL"`
	Title        string   `help:"New poll title" short:"t"`
	AddOption    []string `help:"Add option (repeatable)" short:"a"`
	RemoveOption []int    `help:"Remove option by position index (repeatable)" short:"r"`
}

// Run updates a ranking poll via the API.
func (c *RankingUpdateCmd) Run(flags *RootFlags) error {
	if c.Title == "" && len(c.AddOption) == 0 && len(c.RemoveOption) == 0 {
		return fmt.Errorf("specify at least one field to update")
	}

	id := api.ParsePollID(c.ID)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	req := &api.UpdatePollRequest{}

	if c.Title != "" {
		req.Title = c.Title
	}

	// If removing options, fetch existing poll first
	if len(c.RemoveOption) > 0 {
		poll, err := client.GetPoll(context.Background(), id)
		if err != nil {
			return fmt.Errorf("fetch poll for option removal: %w", err)
		}

		removeSet := make(map[int]bool, len(c.RemoveOption))
		for _, idx := range c.RemoveOption {
			removeSet[idx] = true
		}

		var kept []*api.PollOption
		for _, opt := range poll.PollOptions {
			if !removeSet[opt.Position] {
				kept = append(kept, opt)
			}
		}

		// Renumber positions
		for i, opt := range kept {
			opt.Position = i
		}

		req.PollOptions = kept
	}

	// Append new options
	for _, val := range c.AddOption {
		pos := len(req.PollOptions)
		req.PollOptions = append(req.PollOptions, &api.PollOption{
			Type:     api.OptionTypeText,
			Value:    val,
			Position: pos,
		})
	}

	poll, err := client.UpdatePoll(context.Background(), id, req)
	if err != nil {
		return err
	}

	pollURL := pollBaseURL + poll.ID

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "Type", "URL", "Options", "Votes"}
	rows := [][]string{{
		poll.ID,
		poll.Title,
		poll.Type,
		pollURL,
		fmt.Sprintf("%d", len(poll.PollOptions)),
		voteCount(poll),
	}}

	return f.Output(poll, headers, rows)
}
