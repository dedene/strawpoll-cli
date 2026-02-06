package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

// MeetingCreateCmd creates a meeting poll with date/time options.
type MeetingCreateCmd struct {
	Title string `arg:"" required:"" help:"Meeting poll title"`

	// Date/Time options
	Date  []string `help:"All-day date in YYYY-MM-DD format (repeatable)" short:"d"`
	Range []string `help:"Time range as 'YYYY-MM-DD HH:MM-HH:MM' (repeatable)" short:"r"`
	Tz    string   `help:"IANA timezone (e.g. Europe/Berlin); defaults to local" default:""`

	// Meeting details
	Location    string `help:"Meeting location" default:""`
	Description string `help:"Poll description" default:""`

	// Meeting-specific options
	AllowMaybe bool   `help:"Allow 'if need be' responses" default:"true" negatable:""`
	Dupcheck   string `help:"Duplication checking: ip, session, none" default:"none"`
	Deadline   string `help:"Deadline (RFC3339 or duration like 24h)" default:""`
}

// Run creates a meeting poll via the API.
func (c *MeetingCreateCmd) Run(flags *RootFlags) error {
	if len(c.Date) == 0 && len(c.Range) == 0 {
		// No dates/ranges provided — try interactive wizard
		if !tui.IsInteractive() {
			return fmt.Errorf("specify at least one --date or --range; or run interactively in a terminal")
		}
		return c.runWizard(flags)
	}

	return c.createFromFlags(flags)
}

// runWizard launches the interactive meeting wizard and maps results to command fields.
func (c *MeetingCreateCmd) runWizard(flags *RootFlags) error {
	result, err := tui.RunMeetingWizard()
	if err != nil {
		return err
	}

	// Map wizard results to command fields
	c.Date = result.Dates
	c.Range = result.TimeRanges
	if result.Timezone != "" {
		c.Tz = result.Timezone
	}
	c.Location = result.Location
	c.Description = result.Description
	c.AllowMaybe = result.AllowMaybe
	c.Dupcheck = result.Dupcheck

	return c.createFromFlags(flags)
}

// createFromFlags handles the flag-based meeting creation flow.
func (c *MeetingCreateCmd) createFromFlags(flags *RootFlags) error {
	loc, err := c.resolveTimezone()
	if err != nil {
		return err
	}

	var options []*api.PollOption

	pos := 0

	// Parse --date values first
	for _, d := range c.Date {
		opt, err := parseDateOption(d)
		if err != nil {
			return fmt.Errorf("invalid --date %q: %w", d, err)
		}

		opt.Position = pos
		pos++

		options = append(options, opt)
	}

	// Parse --range values
	for _, r := range c.Range {
		opt, err := parseTimeRange(r, loc)
		if err != nil {
			return fmt.Errorf("invalid --range %q: %w", r, err)
		}

		opt.Position = pos
		pos++

		options = append(options, opt)
	}

	req := c.buildRequest(options, loc)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}

	defer client.Close()

	poll, err := client.CreatePoll(context.Background(), req)
	if err != nil {
		return err
	}

	pollURL := pollBaseURL + poll.ID

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	if err := f.OutputSingle(poll, [][2]string{
		{"ID", poll.ID},
		{"Title", poll.Title},
		{"URL", pollURL},
	}); err != nil {
		return err
	}

	if flags.Copy {
		if err := clipboard.WriteAll(pollURL); err != nil {
			fmt.Fprintf(os.Stderr, "clipboard: %v\n", err)
		}
	}

	if flags.Open {
		if err := browser.OpenURL(pollURL); err != nil {
			fmt.Fprintf(os.Stderr, "browser: %v\n", err)
		}
	}

	return nil
}

func (c *MeetingCreateCmd) resolveTimezone() (*time.Location, error) {
	if c.Tz != "" {
		loc, err := time.LoadLocation(c.Tz)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", c.Tz, err)
		}

		return loc, nil
	}

	return time.Now().Location(), nil
}

func (c *MeetingCreateCmd) buildRequest(options []*api.PollOption, loc *time.Location) *api.CreatePollRequest {
	pollCfg := &api.PollConfig{
		VoteType:            api.VoteTypeParticipantGrid,
		IsMultipleChoice:    boolPtr(true),
		RequireVoterNames:   boolPtr(true),
		AllowIndeterminate:  boolPtr(c.AllowMaybe),
		DuplicationChecking: c.Dupcheck,
		EditVotePermissions: "admin_voter",
	}

	if c.Deadline != "" {
		pollCfg.DeadlineAt = parseDeadlineUnix(c.Deadline)
	}

	tzName := ""
	if c.Tz != "" {
		tzName = c.Tz
	} else {
		tzName = loc.String()
	}

	pollMeta := &api.PollMeta{
		Description: c.Description,
		Location:    c.Location,
		Timezone:    tzName,
	}

	return &api.CreatePollRequest{
		Title:       c.Title,
		Type:        api.PollTypeMeeting,
		PollOptions: options,
		PollConfig:  pollCfg,
		PollMeta:    pollMeta,
	}
}

// parseDateOption validates a YYYY-MM-DD date and creates a date PollOption.
func parseDateOption(s string) (*api.PollOption, error) {
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("expected YYYY-MM-DD format: %w", err)
	}

	return &api.PollOption{
		Type:  api.OptionTypeDate,
		Value: s,
		Date:  s,
	}, nil
}

// parseTimeRange parses "YYYY-MM-DD HH:MM-HH:MM" or "YYYY-MM-DD HH:MM"
// (open-ended) into a time_range PollOption with Unix timestamps.
func parseTimeRange(s string, loc *time.Location) (*api.PollOption, error) {
	// Split into date and time parts: "2024-08-13 10:00-11:00"
	// At least "YYYY-MM-DD HH:MM" = date(10) + space(1) + time(5) = 16 chars
	parts := strings.SplitN(s, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected 'YYYY-MM-DD HH:MM-HH:MM' or 'YYYY-MM-DD HH:MM'")
	}

	datePart := parts[0]
	timePart := parts[1]

	// Validate the date portion
	if _, err := time.Parse("2006-01-02", datePart); err != nil {
		return nil, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", datePart)
	}

	// Split time part on "-" for start-end
	timeParts := strings.SplitN(timePart, "-", 2)

	startStr := datePart + " " + timeParts[0]

	startTime, err := time.ParseInLocation("2006-01-02 15:04", startStr, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid start time %q: expected HH:MM", timeParts[0])
	}

	startUnix := startTime.Unix()

	opt := &api.PollOption{
		Type:      api.OptionTypeTimeRange,
		Value:     s,
		StartTime: &startUnix,
	}

	// End time is optional (open-ended range)
	if len(timeParts) == 2 && timeParts[1] != "" {
		endStr := datePart + " " + timeParts[1]

		endTime, err := time.ParseInLocation("2006-01-02 15:04", endStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid end time %q: expected HH:MM", timeParts[1])
		}

		endUnix := endTime.Unix()
		opt.EndTime = &endUnix
	}

	return opt, nil
}
