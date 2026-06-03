package theme

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DefaultSlug is the theme used when none is specified.
const DefaultSlug = "gum"

// Theme holds the resolved lipgloss styles for each UI element.
// No style sets a background color — the terminal background is always used.
type Theme struct {
	Cursor      lipgloss.Style // ">" prefix in the list
	Selected    lipgloss.Style // highlighted item label + preview title
	Item        lipgloss.Style // normal item label
	Header      lipgloss.Style // menu-level title
	Description lipgloss.Style // preview body text
	Border      lipgloss.Style // separators (│, ─)
	Help        lipgloss.Style // key hint line
	NoColor     bool
}

// New builds a Theme from a palette slug and the noColor flag.
// Slug resolution order: slug arg → NO_COLOR env → gum fallback.
// Returns plain unstyled output when noColor is true or NO_COLOR is set.
func New(slug string, noColor bool) *Theme {
	if _, noColorSet := os.LookupEnv("NO_COLOR"); noColor || noColorSet {
		return &Theme{NoColor: true}
	}

	p := Lookup(strings.ToLower(strings.TrimSpace(slug)))

	return &Theme{
		Cursor:      lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Item:        lipgloss.NewStyle().Foreground(p.Fg),
		Header:      lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Description: lipgloss.NewStyle().Foreground(p.Fg),
		Border:      lipgloss.NewStyle().Foreground(p.Muted),
		Help:        lipgloss.NewStyle().Foreground(p.Muted),
	}
}
