package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderTable writes a formatted table to w.
// If colors are not enabled, falls back to SimpleTable.
// If colors enabled, uses lipgloss/table with styled headers and alternating rows.
func RenderTable(w io.Writer, headers []string, rows [][]string, colors *Colors) error {
	if colors == nil || !colors.Enabled() {
		return SimpleTable(w, headers, rows)
	}

	// No data to render
	if len(headers) == 0 && len(rows) == 0 {
		return nil
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60a5fa")).
		PaddingRight(2)

	defaultStyle := lipgloss.NewStyle().PaddingRight(2)
	faintStyle := lipgloss.NewStyle().Faint(true).PaddingRight(2)

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(true).
		BorderColumn(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row%2 == 0 {
				return defaultStyle
			}
			return faintStyle
		})

	_, err := fmt.Fprintln(w, t.Render())
	return err
}

// SimpleTable writes a plain table using text/tabwriter (no-color fallback).
func SimpleTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)

	if len(headers) > 0 {
		if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
			return err
		}
		// Separator
		sep := make([]string, len(headers))
		for i, h := range headers {
			sep[i] = strings.Repeat("-", len(h))
		}
		if _, err := fmt.Fprintln(tw, strings.Join(sep, "\t")); err != nil {
			return err
		}
	}

	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}
