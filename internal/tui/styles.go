package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// StderrRenderer is a lipgloss renderer targeting stderr.
// Using stderr avoids color profile mismatch when TUI output goes to stderr
// while data goes to stdout.
var StderrRenderer = lipgloss.NewRenderer(os.Stderr)

// Shared styles for consistent look across TUI components.
var (
	TitleStyle    = StderrRenderer.NewStyle().Bold(true)
	SubtitleStyle = StderrRenderer.NewStyle().Faint(true)
	SelectedStyle = StderrRenderer.NewStyle().Foreground(lipgloss.Color("2"))
	ErrorStyle    = StderrRenderer.NewStyle().Foreground(lipgloss.Color("1"))
)
