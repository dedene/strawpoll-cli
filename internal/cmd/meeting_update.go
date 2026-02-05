package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// MeetingUpdateCmd updates an existing meeting poll.
type MeetingUpdateCmd struct {
	ID       string   `arg:"" required:"" help:"Meeting poll ID or URL"`
	Title    string   `help:"New poll title" short:"t"`
	Location string   `help:"Meeting location"`
	AddDate  []string `help:"Add all-day date YYYY-MM-DD (repeatable)" short:"d"`
	AddRange []string `help:"Add time range 'YYYY-MM-DD HH:MM-HH:MM' (repeatable)" short:"r"`
	Tz       string   `help:"IANA timezone (e.g. Europe/Berlin)"`
}

// Run updates a meeting poll via the API.
func (c *MeetingUpdateCmd) Run(flags *RootFlags) error {
	if c.Title == "" && c.Location == "" && c.Tz == "" && len(c.AddDate) == 0 && len(c.AddRange) == 0 {
		return fmt.Errorf("specify at least one field to update")
	}

	id := api.ParsePollID(c.ID)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()
	req := &api.UpdatePollRequest{}

	if c.Title != "" {
		req.Title = c.Title
	}

	// Update PollMeta if location or timezone changed.
	if c.Location != "" || c.Tz != "" {
		req.PollMeta = &api.PollMeta{}

		if c.Location != "" {
			req.PollMeta.Location = c.Location
		}

		if c.Tz != "" {
			if _, err := time.LoadLocation(c.Tz); err != nil {
				return fmt.Errorf("invalid timezone %q: %w", c.Tz, err)
			}

			req.PollMeta.Timezone = c.Tz
		}
	}

	// Add new date/range options -- fetch existing to determine next position.
	if len(c.AddDate) > 0 || len(c.AddRange) > 0 {
		poll, err := client.GetPoll(ctx, id)
		if err != nil {
			return fmt.Errorf("fetch poll for option additions: %w", err)
		}

		// Resolve timezone for time range parsing.
		loc := resolveUpdateTimezone(c.Tz, poll)

		// Start from existing options.
		opts := poll.PollOptions
		pos := len(opts)

		for _, d := range c.AddDate {
			opt, err := parseDateOption(d)
			if err != nil {
				return fmt.Errorf("invalid --add-date %q: %w", d, err)
			}

			opt.Position = pos
			pos++

			opts = append(opts, opt)
		}

		for _, r := range c.AddRange {
			opt, err := parseTimeRange(r, loc)
			if err != nil {
				return fmt.Errorf("invalid --add-range %q: %w", r, err)
			}

			opt.Position = pos
			pos++

			opts = append(opts, opt)
		}

		req.PollOptions = opts
	}

	poll, err := client.UpdatePoll(ctx, id, req)
	if err != nil {
		return err
	}

	pollURL := pollBaseURL + poll.ID
	loc := meetingLocation(poll)

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "Location", "Timezone", "Options", "Votes"}
	rows := [][]string{{
		poll.ID,
		poll.Title,
		meetingLocationStr(poll),
		meetingTimezoneStr(poll),
		formatMeetingOptions(poll, loc),
		voteCount(poll),
	}}

	fmt.Fprintf(os.Stderr, "Meeting poll updated: %s\n", pollURL)

	return f.Output(poll, headers, rows)
}

// resolveUpdateTimezone picks timezone for parsing new time ranges:
// explicit --tz flag > poll's existing timezone > UTC.
func resolveUpdateTimezone(tz string, poll *api.Poll) *time.Location {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}

	return meetingLocation(poll)
}
