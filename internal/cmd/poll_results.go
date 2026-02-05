package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/auth"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// PollResultsCmd displays poll results.
type PollResultsCmd struct {
	ID           string `arg:"" required:"" help:"Poll ID or URL"`
	Participants bool   `help:"Show per-participant breakdown" short:"p"`
}

// Run fetches and displays poll results.
func (c *PollResultsCmd) Run(flags *RootFlags) error {
	id := api.ParsePollID(c.ID)

	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := api.NewClient(apiKey)
	defer client.Close()

	results, err := client.GetPollResults(context.Background(), id)
	if err != nil {
		return err
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)

	headers, rows := resultsTable(results)
	if err := f.Output(results, headers, rows); err != nil {
		return err
	}

	if c.Participants && len(results.PollParticipants) > 0 {
		fmt.Fprintln(os.Stdout)

		pHeaders, pRows := participantsTable(results)
		if err := f.Output(results.PollParticipants, pHeaders, pRows); err != nil {
			return err
		}
	}

	return nil
}

func resultsTable(r *api.PollResults) ([]string, [][]string) {
	total := r.VoteCount
	headers := []string{"Option", "Votes", "Percentage"}
	rows := make([][]string, 0, len(r.PollOptions))

	for _, opt := range r.PollOptions {
		pct := 0.0
		if total > 0 {
			pct = float64(opt.VoteCount) / float64(total) * 100
		}

		rows = append(rows, []string{
			opt.Value,
			fmt.Sprintf("%d", opt.VoteCount),
			fmt.Sprintf("%.1f%%", pct),
		})
	}

	return headers, rows
}

func participantsTable(r *api.PollResults) ([]string, [][]string) {
	// Build header: Name + each option value
	headers := make([]string, 0, 1+len(r.PollOptions))
	headers = append(headers, "Name")

	for _, opt := range r.PollOptions {
		headers = append(headers, opt.Value)
	}

	rows := make([][]string, 0, len(r.PollParticipants))

	for _, p := range r.PollParticipants {
		row := make([]string, 0, 1+len(r.PollOptions))
		name := p.Name
		if name == "" {
			name = "Anonymous"
		}

		row = append(row, name)

		// PollVotes contains pointers to option indices the participant voted for
		voted := voteSet(p.PollVotes)

		for i := range r.PollOptions {
			if voted[i] {
				row = append(row, "x")
			} else {
				row = append(row, strings.Repeat(" ", 1))
			}
		}

		rows = append(rows, row)
	}

	return headers, rows
}

func voteSet(votes []*int) map[int]bool {
	m := make(map[int]bool, len(votes))

	for _, v := range votes {
		if v != nil {
			m[*v] = true
		}
	}

	return m
}
