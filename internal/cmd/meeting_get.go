package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// MeetingGetCmd retrieves meeting poll details.
type MeetingGetCmd struct {
	ID string `arg:"" required:"" help:"Meeting poll ID or URL"`
}

// Run fetches and displays a meeting poll with formatted timeslots.
func (c *MeetingGetCmd) Run(flags *RootFlags) error {
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

	loc := meetingLocation(poll)

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	return f.OutputSingle(poll, [][2]string{
		{"ID", poll.ID},
		{"Title", poll.Title},
		{"Location", meetingLocationStr(poll)},
		{"Timezone", meetingTimezoneStr(poll)},
		{"Options", formatMeetingOptions(poll, loc)},
		{"Votes", voteCount(poll)},
	})
}

// meetingLocation returns the *time.Location for the poll timezone.
func meetingLocation(p *api.Poll) *time.Location {
	if p.PollMeta != nil && p.PollMeta.Timezone != "" {
		if loc, err := time.LoadLocation(p.PollMeta.Timezone); err == nil {
			return loc
		}
	}

	return time.UTC
}

func meetingLocationStr(p *api.Poll) string {
	if p.PollMeta != nil && p.PollMeta.Location != "" {
		return p.PollMeta.Location
	}

	return "-"
}

func meetingTimezoneStr(p *api.Poll) string {
	if p.PollMeta != nil && p.PollMeta.Timezone != "" {
		return p.PollMeta.Timezone
	}

	return "UTC"
}

// formatMeetingOptions returns a compact summary of meeting timeslots.
func formatMeetingOptions(p *api.Poll, loc *time.Location) string {
	n := len(p.PollOptions)
	if n == 0 {
		return "0"
	}

	// Show count + first option as preview
	first := formatMeetingOption(p.PollOptions[0], loc)
	if n == 1 {
		return first
	}

	return fmt.Sprintf("%s (+%d more)", first, n-1)
}

// formatMeetingOption returns a human-readable time slot string.
func formatMeetingOption(opt *api.PollOption, loc *time.Location) string {
	switch opt.Type {
	case api.OptionTypeDate:
		if opt.Date != "" {
			if t, err := time.Parse("2006-01-02", opt.Date); err == nil {
				return t.Format("Mon Jan 2")
			}
		}

		return opt.Value
	case api.OptionTypeTimeRange:
		if opt.StartTime != nil && opt.EndTime != nil {
			start := time.Unix(*opt.StartTime, 0).In(loc)
			end := time.Unix(*opt.EndTime, 0).In(loc)

			return fmt.Sprintf("%s %s-%s",
				start.Format("Mon Jan 2"),
				start.Format("15:04"),
				end.Format("15:04"),
			)
		}

		return opt.Value
	default:
		return opt.Value
	}
}
