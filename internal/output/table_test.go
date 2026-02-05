package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTable_ColorsDisabled(t *testing.T) {
	colors := &Colors{enabled: false}
	headers := []string{"Name", "Votes", "Pct"}
	rows := [][]string{
		{"Option A", "10", "50%"},
		{"Option B", "10", "50%"},
	}

	var buf bytes.Buffer
	err := RenderTable(&buf, headers, rows, colors)
	if err != nil {
		t.Fatalf("RenderTable error: %v", err)
	}

	output := buf.String()

	// Should contain headers
	if !strings.Contains(output, "Name") {
		t.Error("expected output to contain header 'Name'")
	}
	if !strings.Contains(output, "Votes") {
		t.Error("expected output to contain header 'Votes'")
	}

	// Should contain data
	if !strings.Contains(output, "Option A") {
		t.Error("expected output to contain 'Option A'")
	}
	if !strings.Contains(output, "Option B") {
		t.Error("expected output to contain 'Option B'")
	}

	// Should have separator line
	if !strings.Contains(output, "----") {
		t.Error("expected output to contain separator dashes")
	}
}

func TestRenderTable_NilColors(t *testing.T) {
	headers := []string{"ID", "Title"}
	rows := [][]string{{"1", "Test"}}

	var buf bytes.Buffer
	err := RenderTable(&buf, headers, rows, nil)
	if err != nil {
		t.Fatalf("RenderTable with nil colors error: %v", err)
	}

	if !strings.Contains(buf.String(), "Test") {
		t.Error("expected output to contain data")
	}
}

func TestRenderTable_EmptyRows(t *testing.T) {
	colors := &Colors{enabled: false}

	var buf bytes.Buffer
	err := RenderTable(&buf, []string{"Name"}, nil, colors)
	if err != nil {
		t.Fatalf("RenderTable empty rows error: %v", err)
	}

	// Should still have header
	if !strings.Contains(buf.String(), "Name") {
		t.Error("expected header in output")
	}
}

func TestRenderTable_EmptyHeadersAndRows(t *testing.T) {
	colors := &Colors{enabled: true}

	var buf bytes.Buffer
	err := RenderTable(&buf, nil, nil, colors)
	if err != nil {
		t.Fatalf("RenderTable empty error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestSimpleTable(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		headers := []string{"ID", "Name"}
		rows := [][]string{
			{"1", "Alice"},
			{"2", "Bob"},
		}

		var buf bytes.Buffer
		err := SimpleTable(&buf, headers, rows)
		if err != nil {
			t.Fatalf("SimpleTable error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 4 { // header + separator + 2 rows
			t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
		}
	})

	t.Run("no headers", func(t *testing.T) {
		rows := [][]string{{"a", "b"}, {"c", "d"}}

		var buf bytes.Buffer
		err := SimpleTable(&buf, nil, rows)
		if err != nil {
			t.Fatalf("SimpleTable error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
	})

	t.Run("empty data", func(t *testing.T) {
		var buf bytes.Buffer
		err := SimpleTable(&buf, nil, nil)
		if err != nil {
			t.Fatalf("SimpleTable error: %v", err)
		}
		// Should produce no output (or minimal whitespace)
	})
}
