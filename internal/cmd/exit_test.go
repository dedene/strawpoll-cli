package cmd

import (
	"errors"
	"testing"
)

func TestExitCode_Nil(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Errorf("ExitCode(nil) = %d, want 0", got)
	}
}

func TestExitCode_GenericError(t *testing.T) {
	if got := ExitCode(errors.New("boom")); got != 1 {
		t.Errorf("ExitCode(generic) = %d, want 1", got)
	}
}

func TestExitCode_ExitError(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"success", CodeSuccess, 0},
		{"error", CodeError, 1},
		{"usage", CodeUsage, 2},
		{"auth", CodeAuth, 3},
		{"api", CodeAPI, 4},
		{"rate-limit", CodeRateLimit, 5},
		{"negative", -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ExitError{Code: tt.code, Err: errors.New("test")}
			if got := ExitCode(err); got != tt.want {
				t.Errorf("ExitCode(code=%d) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestExitError_Error(t *testing.T) {
	e := &ExitError{Code: 1, Err: errors.New("test error")}
	if got := e.Error(); got != "test error" {
		t.Errorf("Error() = %q, want %q", got, "test error")
	}
}

func TestExitError_ErrorNilErr(t *testing.T) {
	e := &ExitError{Code: 1, Err: nil}
	if got := e.Error(); got != "" {
		t.Errorf("Error() = %q, want empty", got)
	}
}

func TestExitError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	e := &ExitError{Code: 1, Err: inner}
	if got := e.Unwrap(); got != inner {
		t.Errorf("Unwrap() = %v, want %v", got, inner)
	}
}
