package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// RankingListCmd lists the user's ranking polls.
type RankingListCmd struct {
	Limit int `help:"Max results per page" default:"20" short:"l"`
	Page  int `help:"Page number" default:"1" short:"p"`
}

// Run lists ranking polls with client-side type filtering.
func (c *RankingListCmd) Run(flags *RootFlags) error {
	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.ListMyPolls(context.Background(), "created", c.Page, c.Limit)
	if err != nil {
		return err
	}

	// Filter to ranking polls only
	var rankings []api.Poll
	for _, p := range resp.Data {
		if p.Type == api.PollTypeRanking {
			rankings = append(rankings, p)
		}
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "Options", "Votes", "Created"}

	rows := make([][]string, 0, len(rankings))
	for _, p := range rankings {
		votes := "0"
		if p.PollMeta != nil {
			votes = fmt.Sprintf("%d", p.PollMeta.VoteCount)
		}

		rows = append(rows, []string{
			p.ID,
			p.Title,
			fmt.Sprintf("%d", len(p.PollOptions)),
			votes,
			time.Unix(p.CreatedAt, 0).Format("2006-01-02"),
		})
	}

	if err := f.Output(rankings, headers, rows); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Page %d of %d (%d total ranking polls shown)\n",
		resp.Pagination.Page, (resp.Pagination.Total+resp.Pagination.Limit-1)/resp.Pagination.Limit, len(rankings))

	return nil
}
