package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// Confirm displays an interactive confirmation prompt on stderr.
// Returns (false, error) if stdin is not a TTY — callers should use --force
// in non-interactive mode.
func Confirm(prompt string) (bool, error) {
	if !IsInteractive() {
		return false, fmt.Errorf("confirmation required; use --force in non-interactive mode")
	}

	var confirmed bool

	confirm := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed)

	err := huh.NewForm(huh.NewGroup(confirm)).
		WithProgramOptions(tea.WithOutput(os.Stderr)).
		Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}
