package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// MeetingListCmd lists the user's meeting polls.
type MeetingListCmd struct {
	Limit int `help:"Results per page" default:"20"`
	Page  int `help:"Page number" default:"1"`
}

// Run lists meeting polls with client-side type filtering and pagination.
func (c *MeetingListCmd) Run(flags *RootFlags) error {
	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.ListMyPolls(context.Background(), "created", c.Page, c.Limit)
	if err != nil {
		return err
	}

	// Filter to meeting polls client-side.
	var meetings []api.Poll
	for _, p := range resp.Data {
		if p.Type == api.PollTypeMeeting {
			meetings = append(meetings, p)
		}
	}

	if len(meetings) == 0 {
		fmt.Fprintln(os.Stderr, "No meeting polls found.")

		return nil
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "Location", "Options", "Votes", "Created"}
	rows := make([][]string, 0, len(meetings))

	for i := range meetings {
		p := &meetings[i]
		rows = append(rows, []string{
			p.ID,
			p.Title,
			meetingLocationStr(p),
			fmt.Sprintf("%d", len(p.PollOptions)),
			voteCount(p),
			time.Unix(p.CreatedAt, 0).Format("2006-01-02"),
		})
	}

	if err := f.Output(meetings, headers, rows); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Page %d/%d (%d meeting polls shown)\n",
		resp.Pagination.Page,
		paginationPages(resp.Pagination),
		len(meetings),
	)

	return nil
}

// paginationPages calculates total pages from pagination metadata.
func paginationPages(p api.Pagination) int {
	if p.Limit <= 0 {
		return 1
	}

	pages := p.Total / p.Limit
	if p.Total%p.Limit > 0 {
		pages++
	}

	if pages < 1 {
		return 1
	}

	return pages
}
