package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// MeetingWizardResult holds the collected meeting wizard inputs.
type MeetingWizardResult struct {
	Dates       []string // YYYY-MM-DD format
	TimeRanges  []string // "YYYY-MM-DD HH:MM-HH:MM" format
	Timezone    string   // IANA timezone name
	Location    string
	Description string
	AllowMaybe  bool
	Dupcheck    string
}

// RunMeetingWizard launches the interactive meeting creation wizard.
// Renders on stderr; returns collected inputs for the meeting create flow.
func RunMeetingWizard() (*MeetingWizardResult, error) {
	var (
		datesText  string
		rangesText string
		timezone   string
		location   string
		description string
		allowMaybe  = true
		dupcheck    = "none"
	)

	form := huh.NewForm(
		// Group 1: Date/Time Selection
		huh.NewGroup(
			huh.NewText().
				Title("All-day dates (one per line, YYYY-MM-DD)").
				Placeholder("2025-03-01\n2025-03-02").
				Lines(5).
				Value(&datesText).
				Validate(validateDateLines),

			huh.NewText().
				Title("Time ranges (one per line, YYYY-MM-DD HH:MM-HH:MM)").
				Placeholder("2025-03-01 10:00-11:00\n2025-03-01 14:00-15:00").
				Lines(5).
				Value(&rangesText).
				Validate(validateRangeLines),
		).Title("Date & Time Options"),

		// Group 2: Timezone, Location, Description
		huh.NewGroup(
			huh.NewInput().
				Title("Timezone (IANA, e.g. Europe/Berlin)").
				Placeholder("Local timezone").
				Value(&timezone).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					_, err := time.LoadLocation(s)
					if err != nil {
						return fmt.Errorf("invalid IANA timezone: %s", s)
					}
					return nil
				}),

			huh.NewInput().
				Title("Location").
				Placeholder("optional").
				Value(&location),

			huh.NewText().
				Title("Description").
				Lines(3).
				Placeholder("optional").
				Value(&description),
		).Title("Meeting Details"),

		// Group 3: Settings
		huh.NewGroup(
			huh.NewConfirm().
				Title("Allow 'if need be' responses?").
				Affirmative("Yes").
				Negative("No").
				Value(&allowMaybe),

			huh.NewSelect[string]().
				Title("Duplicate checking").
				Options(
					huh.NewOption("IP address", "ip"),
					huh.NewOption("Browser session", "session"),
					huh.NewOption("None", "none"),
				).
				Value(&dupcheck),
		).Title("Settings"),
	).WithProgramOptions(tea.WithOutput(os.Stderr))

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(parseLinesNonEmpty(datesText)) == 0 && len(parseLinesNonEmpty(rangesText)) == 0 {
		return nil, fmt.Errorf("provide at least one date or time range")
	}

	return &MeetingWizardResult{
		Dates:       parseLinesNonEmpty(datesText),
		TimeRanges:  parseLinesNonEmpty(rangesText),
		Timezone:    timezone,
		Location:    location,
		Description: description,
		AllowMaybe:  allowMaybe,
		Dupcheck:    dupcheck,
	}, nil
}

// validateDateLines validates that each non-empty line is YYYY-MM-DD.
func validateDateLines(s string) error {
	for _, line := range parseLinesNonEmpty(s) {
		if _, err := time.Parse("2006-01-02", line); err != nil {
			return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", line)
		}
	}
	return nil
}

// validateRangeLines validates each non-empty line matches
// "YYYY-MM-DD HH:MM-HH:MM" or "YYYY-MM-DD HH:MM" (open-ended).
func validateRangeLines(s string) error {
	for _, line := range parseLinesNonEmpty(s) {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid range %q: expected 'YYYY-MM-DD HH:MM-HH:MM'", line)
		}
		if _, err := time.Parse("2006-01-02", parts[0]); err != nil {
			return fmt.Errorf("invalid date in range %q: expected YYYY-MM-DD", line)
		}
		timeParts := strings.SplitN(parts[1], "-", 2)
		if _, err := time.Parse("15:04", timeParts[0]); err != nil {
			return fmt.Errorf("invalid start time in range %q: expected HH:MM", line)
		}
		if len(timeParts) == 2 && timeParts[1] != "" {
			if _, err := time.Parse("15:04", timeParts[1]); err != nil {
				return fmt.Errorf("invalid end time in range %q: expected HH:MM", line)
			}
		}
	}
	return nil
}

// parseLinesNonEmpty splits text on newlines, trims whitespace,
// and returns only non-empty lines.
func parseLinesNonEmpty(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
