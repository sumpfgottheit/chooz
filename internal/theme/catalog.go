package theme

import "github.com/charmbracelet/lipgloss"

// Palette holds the colors needed to style chooz's UI.
// Bg is intentionally absent — the terminal background is always used.
type Palette struct {
	Slug    string
	Name    string
	Variant string // "dark", "light", or "adaptive"
	Fg      lipgloss.TerminalColor // primary text (NoColor = terminal default)
	Muted   lipgloss.TerminalColor // help hints, borders, description body
	Primary lipgloss.TerminalColor // cursor, selected item, header
}

func c(s string) lipgloss.Color { return lipgloss.Color(s) }
func a(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

// catalog lists all built-in themes in presentation order (dark → light → adaptive).
var catalog = []Palette{
	// ── dark ────────────────────────────────────────────────────────────────
	{
		Slug: "gum", Name: "Gum (charm.sh)", Variant: "dark",
		Fg: lipgloss.NoColor{}, Muted: c("240"), Primary: c("212"),
	},
	{
		Slug: "gruvbox-dark", Name: "Gruvbox Dark", Variant: "dark",
		Fg: c("#ebdbb2"), Muted: c("#928374"), Primary: c("#fe8019"),
	},
	{
		Slug: "nord", Name: "Nord", Variant: "dark",
		Fg: c("#d8dee9"), Muted: c("#4c566a"), Primary: c("#88c0d0"),
	},
	{
		Slug: "dracula", Name: "Dracula", Variant: "dark",
		Fg: c("#f8f8f2"), Muted: c("#6272a4"), Primary: c("#bd93f9"),
	},
	{
		Slug: "catppuccin-mocha", Name: "Catppuccin Mocha", Variant: "dark",
		Fg: c("#cdd6f4"), Muted: c("#6c7086"), Primary: c("#cba6f7"),
	},
	{
		Slug: "tokyo-night", Name: "Tokyo Night", Variant: "dark",
		Fg: c("#c0caf5"), Muted: c("#565f89"), Primary: c("#7aa2f7"),
	},
	// ── light ───────────────────────────────────────────────────────────────
	{
		Slug: "gruvbox-light", Name: "Gruvbox Light", Variant: "light",
		Fg: c("#3c3836"), Muted: c("#7c6f64"), Primary: c("#af3a03"),
	},
	{
		Slug: "solarized-light", Name: "Solarized Light", Variant: "light",
		Fg: c("#657b83"), Muted: c("#93a1a1"), Primary: c("#268bd2"),
	},
	{
		Slug: "catppuccin-latte", Name: "Catppuccin Latte", Variant: "light",
		Fg: c("#4c4f69"), Muted: c("#8c8fa1"), Primary: c("#8839ef"),
	},
	{
		Slug: "one-light", Name: "One Light", Variant: "light",
		Fg: c("#383a42"), Muted: c("#a0a1a7"), Primary: c("#4078f2"),
	},
	{
		Slug: "rose-pine-dawn", Name: "Rosé Pine Dawn", Variant: "light",
		Fg: c("#575279"), Muted: c("#9893a5"), Primary: c("#907aa9"),
	},
	// ── adaptive ────────────────────────────────────────────────────────────
	{
		Slug: "solarized", Name: "Solarized (adaptive)", Variant: "adaptive",
		Fg:      a("#657b83", "#839496"),
		Muted:   a("#93a1a1", "#586e75"),
		Primary: a("#268bd2", "#268bd2"),
	},
	{
		Slug: "everforest", Name: "Everforest (adaptive)", Variant: "adaptive",
		Fg:      a("#5c6a72", "#d3c6aa"),
		Muted:   a("#939f91", "#859289"),
		Primary: a("#8da101", "#a7c080"),
	},
	{
		Slug: "rose-pine", Name: "Rosé Pine (adaptive)", Variant: "adaptive",
		Fg:      a("#575279", "#e0def4"),
		Muted:   a("#9893a5", "#6e6a86"),
		Primary: a("#907aa9", "#c4a7e7"),
	},
	{
		Slug: "kanagawa", Name: "Kanagawa (adaptive)", Variant: "adaptive",
		Fg:      a("#545464", "#dcd7ba"),
		Muted:   a("#8a8980", "#727169"),
		Primary: a("#4d699b", "#7e9cd8"),
	},
	{
		Slug: "base16", Name: "base16 Default (adaptive)", Variant: "adaptive",
		Fg:      a("#383838", "#d8d8d8"),
		Muted:   a("#b8b8b8", "#585858"),
		Primary: a("#7cafc2", "#7cafc2"),
	},
}

// Lookup finds a Palette by slug (case-insensitive). Returns the gum palette
// as fallback when slug is empty or unknown.
func Lookup(slug string) Palette {
	for _, p := range catalog {
		if p.Slug == slug {
			return p
		}
	}
	return catalog[0] // gum
}

// All returns every Palette in catalog order.
func All() []Palette { return catalog }

// Slugs returns all available slugs in catalog order.
func Slugs() []string {
	out := make([]string, len(catalog))
	for i, p := range catalog {
		out[i] = p.Slug
	}
	return out
}
