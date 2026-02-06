package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/output"
)

// PollListCmd lists the user's polls.
type PollListCmd struct {
	Limit int    `help:"Polls per page" default:"20"`
	Page  int    `help:"Page number" default:"1"`
	Type  string `help:"Poll type: created or participated" default:"created" enum:"created,participated"`
}

// Run lists polls via the API.
func (c *PollListCmd) Run(flags *RootFlags) error {
	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.ListMyPolls(context.Background(), c.Type, c.Page, c.Limit)
	if err != nil {
		return err
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "Type", "Votes", "Created"}
	rows := make([][]string, 0, len(resp.Data))

	for _, p := range resp.Data {
		votes := "0"
		if p.PollMeta != nil {
			votes = fmt.Sprintf("%d", p.PollMeta.VoteCount)
		}

		rows = append(rows, []string{
			p.ID,
			p.Title,
			friendlyType(p.Type),
			votes,
			time.Unix(p.CreatedAt, 0).Format("2006-01-02"),
		})
	}

	if err := f.Output(resp, headers, rows); err != nil {
		return err
	}

	totalPages := int(math.Ceil(float64(resp.Pagination.Total) / float64(max(resp.Pagination.Limit, 1))))
	fmt.Fprintf(os.Stderr, "Page %d/%d (%d polls)\n", resp.Pagination.Page, totalPages, resp.Pagination.Total)

	return nil
}

// friendlyType returns a human-friendly label for poll types.
func friendlyType(t string) string {
	switch t {
	case "multiple_choice":
		return "Multiple Choice"
	case "meeting":
		return "Meeting"
	case "ranking":
		return "Ranking"
	default:
		return t
	}
}
