package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/auth"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// PollGetCmd retrieves poll details.
type PollGetCmd struct {
	ID string `arg:"" required:"" help:"Poll ID or URL"`
}

// Run fetches and displays a poll.
func (c *PollGetCmd) Run(flags *RootFlags) error {
	id := api.ParsePollID(c.ID)

	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := api.NewClient(apiKey)
	defer client.Close()

	poll, err := client.GetPoll(context.Background(), id)
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

func voteCount(p *api.Poll) string {
	if p.PollMeta != nil {
		return fmt.Sprintf("%d", p.PollMeta.VoteCount)
	}

	return "0"
}
