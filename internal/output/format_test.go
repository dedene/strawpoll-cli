package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestModeFromFlags(t *testing.T) {
	tests := []struct {
		name      string
		jsonFlag  bool
		plainFlag bool
		want      Mode
	}{
		{"default is table", false, false, ModeTable},
		{"json flag", true, false, ModeJSON},
		{"plain flag", false, true, ModePlain},
		{"json takes precedence over plain", true, true, ModeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModeFromFlags(tt.jsonFlag, tt.plainFlag)
			if got != tt.want {
				t.Errorf("ModeFromFlags(%v, %v) = %v, want %v", tt.jsonFlag, tt.plainFlag, got, tt.want)
			}
		})
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeTable, "table"},
		{ModeJSON, "json"},
		{ModePlain, "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("Mode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Run("produces valid indented JSON", func(t *testing.T) {
		data := map[string]any{
			"name": "Test Poll",
			"id":   "abc123",
		}

		var buf bytes.Buffer
		err := WriteJSON(&buf, data)
		if err != nil {
			t.Fatalf("WriteJSON error: %v", err)
		}

		// Must be valid JSON
		var parsed map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
		}

		// Must be indented
		if !strings.Contains(buf.String(), "\n  ") {
			t.Error("expected indented JSON output")
		}

		// Verify values roundtrip
		if parsed["name"] != "Test Poll" {
			t.Errorf("name = %v, want Test Poll", parsed["name"])
		}
	})

	t.Run("does not escape HTML", func(t *testing.T) {
		data := map[string]string{"url": "https://example.com?a=1&b=2"}

		var buf bytes.Buffer
		if err := WriteJSON(&buf, data); err != nil {
			t.Fatalf("WriteJSON error: %v", err)
		}

		if strings.Contains(buf.String(), `\u0026`) {
			t.Error("expected HTML not to be escaped")
		}
	})

	t.Run("handles slice input", func(t *testing.T) {
		data := []string{"a", "b", "c"}

		var buf bytes.Buffer
		if err := WriteJSON(&buf, data); err != nil {
			t.Fatalf("WriteJSON error: %v", err)
		}

		var parsed []string
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if len(parsed) != 3 {
			t.Errorf("expected 3 items, got %d", len(parsed))
		}
	})
}

func TestWriteTSV(t *testing.T) {
	t.Run("produces tab-separated output with headers", func(t *testing.T) {
		headers := []string{"Name", "Votes", "Pct"}
		rows := [][]string{
			{"Option A", "10", "50%"},
			{"Option B", "10", "50%"},
		}

		var buf bytes.Buffer
		err := WriteTSV(&buf, headers, rows)
		if err != nil {
			t.Fatalf("WriteTSV error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
		}

		// Header line
		if lines[0] != "Name\tVotes\tPct" {
			t.Errorf("header = %q, want %q", lines[0], "Name\tVotes\tPct")
		}

		// Data rows
		if lines[1] != "Option A\t10\t50%" {
			t.Errorf("row 1 = %q, want %q", lines[1], "Option A\t10\t50%")
		}
	})

	t.Run("handles empty headers", func(t *testing.T) {
		rows := [][]string{{"a", "b"}, {"c", "d"}}

		var buf bytes.Buffer
		err := WriteTSV(&buf, nil, rows)
		if err != nil {
			t.Fatalf("WriteTSV error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines (no header), got %d", len(lines))
		}
	})

	t.Run("handles empty rows", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteTSV(&buf, []string{"H1", "H2"}, nil)
		if err != nil {
			t.Fatalf("WriteTSV error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Fatalf("expected 1 line (header only), got %d", len(lines))
		}
	})
}

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true, false, true)

	if f.Mode != ModeJSON {
		t.Errorf("Mode = %v, want ModeJSON", f.Mode)
	}
	if f.Colors.Enabled() {
		t.Error("expected colors disabled with noColor=true")
	}
}
