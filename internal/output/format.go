// Package output provides output formatting for CLI commands.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Mode represents the output format mode.
type Mode int

const (
	// ModeTable outputs formatted tables (default).
	ModeTable Mode = iota
	// ModeJSON outputs JSON.
	ModeJSON
	// ModePlain outputs tab-separated values.
	ModePlain
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeJSON:
		return "json"
	case ModePlain:
		return "plain"
	default:
		return "table"
	}
}

// ModeFromFlags returns the output mode based on command flags.
// JSON takes precedence over plain.
func ModeFromFlags(jsonFlag, plainFlag bool) Mode {
	if jsonFlag {
		return ModeJSON
	}
	if plainFlag {
		return ModePlain
	}
	return ModeTable
}

// Formatter provides a unified interface for outputting data.
type Formatter struct {
	Writer io.Writer
	Mode   Mode
	Colors *Colors
}

// NewFormatter creates a formatter with the given settings.
func NewFormatter(w io.Writer, jsonFlag, plainFlag, noColor bool) *Formatter {
	return &Formatter{
		Writer: w,
		Mode:   ModeFromFlags(jsonFlag, plainFlag),
		Colors: NewColors(noColor),
	}
}

// Output writes data in the appropriate format.
// For JSON mode, v is encoded directly.
// For table/plain modes, headers and rows are used.
func (f *Formatter) Output(v any, headers []string, rows [][]string) error {
	switch f.Mode {
	case ModeJSON:
		return WriteJSON(f.Writer, v)
	case ModePlain:
		return WriteTSV(f.Writer, headers, rows)
	default:
		return RenderTable(f.Writer, headers, rows, f.Colors)
	}
}

// OutputSingle writes a single resource as key-value pairs (table mode),
// or delegates to JSON/TSV for other modes.
func (f *Formatter) OutputSingle(v any, kvPairs [][2]string) error {
	switch f.Mode {
	case ModeJSON:
		return WriteJSON(f.Writer, v)
	case ModePlain:
		headers := make([]string, len(kvPairs))
		row := make([]string, len(kvPairs))
		for i, kv := range kvPairs {
			headers[i] = kv[0]
			row[i] = kv[1]
		}
		return WriteTSV(f.Writer, headers, [][]string{row})
	default:
		return WriteKV(f.Writer, kvPairs, f.Colors)
	}
}

// WriteKV writes key-value pairs aligned in two columns.
func WriteKV(w io.Writer, pairs [][2]string, colors *Colors) error {
	maxKey := 0
	for _, kv := range pairs {
		if len(kv[0]) > maxKey {
			maxKey = len(kv[0])
		}
	}

	for _, kv := range pairs {
		key := kv[0]
		pad := strings.Repeat(" ", maxKey-len(key))
		if colors != nil && colors.Enabled() {
			key = colors.Bold(key)
		}
		if _, err := fmt.Fprintf(w, "%s%s  %s\n", key, pad, kv[1]); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON writes v as indented JSON to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// WriteTSV writes rows as tab-separated values.
// Headers are written as the first row if non-empty.
func WriteTSV(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) > 0 {
		if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return nil
}
