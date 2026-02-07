package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// MeetingResultsCmd displays meeting poll availability as a timeslot-by-participant grid.
type MeetingResultsCmd struct {
	ID            string `arg:"" required:"" help:"Poll ID or URL"`
	OriginalOrder bool   `help:"Show timeslots in original poll order instead of best availability first" name:"original-order"`
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

	headers, rows, scores := availabilityGrid(poll, results, loc)
	if !c.OriginalOrder {
		sortRowsByScore(rows, scores)
	}

	return f.Output(results, headers, rows)
}

// availabilityGrid builds a timeslot-by-participant grid table.
// Returns headers, rows (one per timeslot), and availability scores for sorting.
func availabilityGrid(poll *api.Poll, results *api.PollResults, loc *time.Location) ([]string, [][]string, []int) {
	// Build headers: "Slot" + participant names + "Total".
	headers := make([]string, 0, 2+len(results.PollParticipants))
	headers = append(headers, "Slot")

	for _, p := range results.PollParticipants {
		name := p.Name
		if name == "" {
			name = "Anonymous"
		}
		headers = append(headers, name)
	}

	headers = append(headers, "Total")

	// Build rows: one per timeslot.
	rows := make([][]string, 0, len(poll.PollOptions))
	scores := make([]int, 0, len(poll.PollOptions))

	for i, opt := range poll.PollOptions {
		row := make([]string, 0, 2+len(results.PollParticipants))
		row = append(row, formatTimeslot(opt, loc))

		for _, p := range results.PollParticipants {
			if i < len(p.PollVotes) {
				row = append(row, voteLabel(p.PollVotes[i]))
			} else {
				row = append(row, "-")
			}
		}

		row = append(row, availabilitySummary(results.PollParticipants, i))
		rows = append(rows, row)
		scores = append(scores, availabilityScore(results.PollParticipants, i))
	}

	return headers, rows, scores
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

// availabilityScore computes a sortable score for a timeslot.
// Yes votes are weighted 1000x more than maybe votes so yes always wins in sort order.
func availabilityScore(participants []*api.PollParticipant, optIdx int) int {
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

	return yes*1000 + maybe
}

// sortRowsByScore sorts rows by their availability scores in descending order.
// Uses stable sort so equal scores preserve original poll order.
func sortRowsByScore(rows [][]string, scores []int) {
	indices := make([]int, len(rows))
	for i := range indices {
		indices[i] = i
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return scores[indices[i]] > scores[indices[j]]
	})

	sorted := make([][]string, len(rows))
	for i, idx := range indices {
		sorted[i] = rows[idx]
	}

	copy(rows, sorted)
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
