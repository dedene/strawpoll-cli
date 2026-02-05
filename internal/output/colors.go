package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Colors provides terminal color support with profile detection.
type Colors struct {
	Profile termenv.Profile
	enabled bool
}

// NewColors creates a Colors instance.
// If noColor is true or NO_COLOR env is set, colors are disabled.
func NewColors(noColor bool) *Colors {
	if noColor || termenv.EnvNoColor() {
		return &Colors{
			Profile: termenv.Ascii,
			enabled: false,
		}
	}

	profile := termenv.EnvColorProfile()
	return &Colors{
		Profile: profile,
		enabled: profile != termenv.Ascii,
	}
}

// IsColorEnabled returns whether color should be enabled given the flag.
func IsColorEnabled(noColor bool) bool {
	if noColor {
		return false
	}
	if termenv.EnvNoColor() {
		return false
	}
	return termenv.EnvColorProfile() != termenv.Ascii
}

// Enabled returns whether colors are enabled.
func (c *Colors) Enabled() bool {
	return c.enabled
}

// Style returns a lipgloss.Style. Returns empty style if colors disabled.
func (c *Colors) Style() lipgloss.Style {
	if !c.enabled {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Renderer(lipgloss.NewRenderer(os.Stdout))
}

// Success returns the string styled as green.
func (c *Colors) Success(s string) string {
	if !c.enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(s)
}

// Error returns the string styled as red.
func (c *Colors) Error(s string) string {
	if !c.enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(s)
}

// Warning returns the string styled as yellow.
func (c *Colors) Warning(s string) string {
	if !c.enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(s)
}

// Dim returns the string in faint style.
func (c *Colors) Dim(s string) string {
	if !c.enabled {
		return s
	}
	return lipgloss.NewStyle().Faint(true).Render(s)
}

// Bold returns the string in bold.
func (c *Colors) Bold(s string) string {
	if !c.enabled {
		return s
	}
	return lipgloss.NewStyle().Bold(true).Render(s)
}
