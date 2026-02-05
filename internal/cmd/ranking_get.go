package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// RankingGetCmd retrieves ranking poll details.
type RankingGetCmd struct {
	ID string `arg:"" required:"" help:"Poll ID or URL"`
}

// Run fetches and displays a ranking poll.
func (c *RankingGetCmd) Run(flags *RootFlags) error {
	id := api.ParsePollID(c.ID)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
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
