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

// Overrides holds optional per-element color values (ANSI-256 or hex).
// Empty strings are ignored. Precedence applied by New(): YAML < env var.
type Overrides struct {
	Cursor, Selected, Item, Header, Description, Border, Help string
}

// New builds a Theme from a palette slug, optional YAML per-element overrides,
// and the noColor flag.
// Precedence: env var > YAML overrides > palette default.
func New(slug string, yaml Overrides, noColor bool) *Theme {
	if _, noColorSet := os.LookupEnv("NO_COLOR"); noColor || noColorSet {
		return &Theme{NoColor: true}
	}

	p := Lookup(strings.ToLower(strings.TrimSpace(slug)))

	thm := &Theme{
		Cursor:      lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Item:        lipgloss.NewStyle().Foreground(p.Fg),
		Header:      lipgloss.NewStyle().Foreground(p.Primary).Bold(true),
		Description: lipgloss.NewStyle().Foreground(p.Fg),
		Border:      lipgloss.NewStyle().Foreground(p.Muted),
		Help:        lipgloss.NewStyle().Foreground(p.Muted),
	}

	// Apply YAML per-element overrides first (lower priority).
	applyOverrides(thm, yaml)

	// Apply env var overrides (higher priority).
	applyOverrides(thm, Overrides{
		Cursor:      os.Getenv("CHOOZ_THEME_CURSOR"),
		Selected:    os.Getenv("CHOOZ_THEME_SELECTED"),
		Item:        os.Getenv("CHOOZ_THEME_ITEM"),
		Header:      os.Getenv("CHOOZ_THEME_HEADER"),
		Description: os.Getenv("CHOOZ_THEME_DESCRIPTION"),
		Border:      os.Getenv("CHOOZ_THEME_BORDER"),
		Help:        os.Getenv("CHOOZ_THEME_HELP"),
	})

	return thm
}

// applyOverrides sets a Foreground color on each element whose override is non-empty.
func applyOverrides(thm *Theme, o Overrides) {
	set := func(s *lipgloss.Style, v string) {
		if v != "" {
			*s = s.Foreground(lipgloss.Color(v))
		}
	}
	set(&thm.Cursor, o.Cursor)
	set(&thm.Selected, o.Selected)
	set(&thm.Item, o.Item)
	set(&thm.Header, o.Header)
	set(&thm.Description, o.Description)
	set(&thm.Border, o.Border)
	set(&thm.Help, o.Help)
}
