package tui

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive returns true when stdin is a terminal (not a pipe/redirect).
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
