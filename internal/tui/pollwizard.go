package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// PollWizardResult holds the collected values from the poll creation wizard.
type PollWizardResult struct {
	Title            string
	Options          []string
	Dupcheck         string
	IsMultipleChoice bool
	IsPrivate        bool
	ResultsVis       string
	AllowComments    bool
}

// RunPollWizard launches a multi-step interactive form for poll creation.
// The form renders on stderr; stdout stays clean for data output.
func RunPollWizard() (*PollWizardResult, error) {
	var (
		title         string
		optionsText   string
		dupcheck      string
		resultsVis    string
		multiChoice   bool
		isPrivate     bool
		allowComments bool
	)

	// Defaults
	dupcheck = "ip"
	resultsVis = "always"

	form := huh.NewForm(
		// Group 1: Title
		huh.NewGroup(
			huh.NewInput().
				Title("Poll title").
				Value(&title).
				Validate(huh.ValidateNotEmpty()),
		),

		// Group 2: Options (one per line)
		huh.NewGroup(
			huh.NewText().
				Title("Options (one per line, min 2)").
				Lines(8).
				Value(&optionsText).
				Validate(validateOptions),
		),

		// Group 3: Settings
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Duplicate checking").
				Options(
					huh.NewOption("IP address", "ip"),
					huh.NewOption("Browser session", "session"),
					huh.NewOption("None", "none"),
				).
				Value(&dupcheck),
			huh.NewSelect[string]().
				Title("Results visibility").
				Options(
					huh.NewOption("Always", "always"),
					huh.NewOption("After deadline", "after_deadline"),
					huh.NewOption("After voting", "after_vote"),
					huh.NewOption("Hidden", "hidden"),
				).
				Value(&resultsVis),
			huh.NewConfirm().
				Title("Allow multiple choices?").
				Value(&multiChoice),
			huh.NewConfirm().
				Title("Private poll?").
				Value(&isPrivate),
			huh.NewConfirm().
				Title("Allow comments?").
				Value(&allowComments),
		),
	).WithProgramOptions(tea.WithOutput(os.Stderr))

	if err := form.Run(); err != nil {
		return nil, err
	}

	options := parseOptionsText(optionsText)

	return &PollWizardResult{
		Title:            title,
		Options:          options,
		Dupcheck:         dupcheck,
		IsMultipleChoice: multiChoice,
		IsPrivate:        isPrivate,
		ResultsVis:       resultsVis,
		AllowComments:    allowComments,
	}, nil
}

// validateOptions ensures at least 2 non-empty lines and at most 30.
func validateOptions(s string) error {
	opts := parseOptionsText(s)
	if len(opts) < 2 {
		return fmt.Errorf("need at least 2 options, got %d", len(opts))
	}

	if len(opts) > 30 {
		return fmt.Errorf("maximum 30 options, got %d", len(opts))
	}

	return nil
}

// parseOptionsText splits text on newlines, trims whitespace, filters empty lines.
func parseOptionsText(s string) []string {
	lines := strings.Split(s, "\n")
	opts := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			opts = append(opts, trimmed)
		}
	}

	return opts
}
