package tui

import (
	"os"
	"testing"
)

func TestIsInteractive_ReturnsBool(t *testing.T) {
	// Verify IsInteractive compiles and returns a bool.
	got := IsInteractive()
	_ = got // type-checked at compile time
}

func TestIsInteractive_NonTTY(t *testing.T) {
	// os.Pipe creates non-TTY file descriptors.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin with the pipe reader temporarily.
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	if IsInteractive() {
		t.Error("IsInteractive() = true for pipe fd; want false")
	}
}
