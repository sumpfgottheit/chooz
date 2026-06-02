package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/saf/chooz/internal/config"
)

// Theme holds the resolved lipgloss styles for each UI element.
type Theme struct {
	Cursor      lipgloss.Style
	Selected    lipgloss.Style
	Item        lipgloss.Style
	Header      lipgloss.Style
	Description lipgloss.Style
	Border      lipgloss.Style
	Help        lipgloss.Style
	NoColor     bool
}

// New builds a Theme from YAML config and flag/env overrides.
// Precedence: env var > YAML theme block > adaptive default.
func New(cfg config.Theme, noColor bool) *Theme {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return &Theme{NoColor: true}
	}

	cursor := resolveColor("CHOOZ_THEME_CURSOR", cfg.Cursor,
		lipgloss.AdaptiveColor{Light: "57", Dark: "212"})
	selected := resolveColor("CHOOZ_THEME_SELECTED", cfg.Selected,
		lipgloss.AdaptiveColor{Light: "57", Dark: "212"})
	item := resolveColor("CHOOZ_THEME_ITEM", cfg.Item,
		lipgloss.AdaptiveColor{Light: "236", Dark: "252"})
	header := resolveColor("CHOOZ_THEME_HEADER", cfg.Header,
		lipgloss.AdaptiveColor{Light: "236", Dark: "252"})
	desc := resolveColor("CHOOZ_THEME_DESCRIPTION", cfg.Description,
		lipgloss.AdaptiveColor{Light: "240", Dark: "246"})
	border := resolveColor("CHOOZ_THEME_BORDER", cfg.Border,
		lipgloss.AdaptiveColor{Light: "250", Dark: "238"})
	help := resolveColor("CHOOZ_THEME_HELP", cfg.Help,
		lipgloss.AdaptiveColor{Light: "244", Dark: "240"})

	return &Theme{
		Cursor:      lipgloss.NewStyle().Foreground(cursor).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(selected).Bold(true),
		Item:        lipgloss.NewStyle().Foreground(item),
		Header:      lipgloss.NewStyle().Foreground(header).Bold(true),
		Description: lipgloss.NewStyle().Foreground(desc),
		Border:      lipgloss.NewStyle().Foreground(border),
		Help:        lipgloss.NewStyle().Foreground(help),
	}
}

func resolveColor(envKey, yamlVal string, fallback lipgloss.TerminalColor) lipgloss.TerminalColor {
	if v := os.Getenv(envKey); v != "" {
		return lipgloss.Color(v)
	}
	if yamlVal != "" {
		return lipgloss.Color(yamlVal)
	}
	return fallback
}
