package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// MeetingResultsCmd displays meeting poll availability as a participant-by-timeslot grid.
type MeetingResultsCmd struct {
	ID string `arg:"" required:"" help:"Poll ID or URL"`
}

// Run fetches meeting poll and results, renders availability grid.
func (c *MeetingResultsCmd) Run(flags *RootFlags) error {
	id := api.ParsePollID(c.ID)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()

	// Fetch both poll (for timezone + option details) and results (for votes).
	poll, err := client.GetPoll(ctx, id)
	if err != nil {
		return err
	}

	results, err := client.GetPollResults(ctx, id)
	if err != nil {
		return err
	}

	// Resolve timezone from poll metadata.
	loc := time.UTC
	if poll.PollMeta != nil && poll.PollMeta.Timezone != "" {
		loc, err = time.LoadLocation(poll.PollMeta.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: unknown timezone %q, using UTC\n", poll.PollMeta.Timezone)
			loc = time.UTC
		}
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)

	headers, rows := availabilityGrid(poll, results, loc)

	return f.Output(results, headers, rows)
}

// availabilityGrid builds the participant-by-timeslot grid table.
// Returns headers (Name + timeslot labels) and rows (participant votes + summary).
func availabilityGrid(poll *api.Poll, results *api.PollResults, loc *time.Location) ([]string, [][]string) {
	// Build headers: "Name" + formatted timeslot for each option.
	headers := make([]string, 0, 1+len(poll.PollOptions))
	headers = append(headers, "Name")

	for _, opt := range poll.PollOptions {
		headers = append(headers, formatTimeslot(opt, loc))
	}

	// Build participant rows.
	rows := make([][]string, 0, len(results.PollParticipants)+1)

	for _, p := range results.PollParticipants {
		row := make([]string, 0, 1+len(poll.PollOptions))

		name := p.Name
		if name == "" {
			name = "Anonymous"
		}
		row = append(row, name)

		for i := range poll.PollOptions {
			if i < len(p.PollVotes) {
				row = append(row, voteLabel(p.PollVotes[i]))
			} else {
				row = append(row, "-")
			}
		}

		rows = append(rows, row)
	}

	// Summary row: count yes+maybe per timeslot.
	summary := make([]string, 0, 1+len(poll.PollOptions))
	summary = append(summary, "Total")

	for i := range poll.PollOptions {
		summary = append(summary, availabilitySummary(results.PollParticipants, i))
	}

	rows = append(rows, summary)

	return headers, rows
}

// formatTimeslot formats a meeting poll option as a human-readable timeslot header.
// - date options: "Mon Jan 2" (e.g. "Tue Aug 13")
// - time_range options: "Mon Jan 2 15:04-15:04" (e.g. "Tue Aug 13 10:00-11:00")
func formatTimeslot(opt *api.PollOption, loc *time.Location) string {
	switch opt.Type {
	case api.OptionTypeDate:
		t, err := time.Parse("2006-01-02", opt.Date)
		if err != nil {
			return opt.Value
		}
		return t.Format("Mon Jan 2")

	case api.OptionTypeTimeRange:
		if opt.StartTime == nil || opt.EndTime == nil {
			return opt.Value
		}

		start := time.Unix(*opt.StartTime, 0).In(loc)
		end := time.Unix(*opt.EndTime, 0).In(loc)

		return start.Format("Mon Jan 2 15:04") + "-" + end.Format("15:04")

	default:
		return opt.Value
	}
}

// voteLabel converts a vote value pointer to a display string.
// 1=Yes, 0=No, 2=Maybe, nil=-.
func voteLabel(v *int) string {
	if v == nil {
		return "-"
	}

	switch *v {
	case 1:
		return "Yes"
	case 0:
		return "No"
	case 2:
		return "Maybe"
	default:
		return fmt.Sprintf("%d", *v)
	}
}

// availabilitySummary counts yes+maybe for a given timeslot index across all participants.
// Returns "{yes}+{maybe}/{total}" or "{yes}/{total}" when maybe is 0.
func availabilitySummary(participants []*api.PollParticipant, optIdx int) string {
	total := len(participants)
	yes, maybe := 0, 0

	for _, p := range participants {
		if optIdx >= len(p.PollVotes) {
			continue
		}

		v := p.PollVotes[optIdx]
		if v == nil {
			continue
		}

		switch *v {
		case 1:
			yes++
		case 2:
			maybe++
		}
	}

	if maybe > 0 {
		return fmt.Sprintf("%d+%d/%d", yes, maybe, total)
	}

	return fmt.Sprintf("%d/%d", yes, total)
}
