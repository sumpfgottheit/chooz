package theme

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// unsetNoColor removes NO_COLOR from the environment for the duration of the
// test, restoring it afterwards. Use instead of t.Setenv("NO_COLOR", "")
// because a set-but-empty NO_COLOR still disables colors per the spec.
func unsetNoColor(t *testing.T) {
	t.Helper()
	prev, wasSet := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if wasSet {
			os.Setenv("NO_COLOR", prev)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	})
}

func TestNoColorFlag(t *testing.T) {
	thm := New("gum", Overrides{}, true)
	if !thm.NoColor {
		t.Error("expected NoColor=true when noColor flag is set")
	}
}

func TestNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	thm := New("gum", Overrides{}, false)
	if !thm.NoColor {
		t.Error("expected NoColor=true when NO_COLOR env is set")
	}
}

func TestNoColorEnvEmpty(t *testing.T) {
	// Per nocolor.org: presence of the variable (even empty) disables colors.
	t.Setenv("NO_COLOR", "")
	thm := New("gum", Overrides{}, false)
	if !thm.NoColor {
		t.Error("expected NoColor=true when NO_COLOR is set (even to empty string)")
	}
}

func TestDefaultTheme(t *testing.T) {
	unsetNoColor(t)
	thm := New("", Overrides{}, false)
	if thm.NoColor {
		t.Error("expected NoColor=false for default theme")
	}
	// Styles should render without panic.
	for _, s := range []string{
		thm.Cursor.Render("x"),
		thm.Selected.Render("x"),
		thm.Item.Render("x"),
		thm.Header.Render("x"),
		thm.Description.Render("x"),
		thm.Border.Render("x"),
		thm.Help.Render("x"),
	} {
		_ = s
	}
}

func TestYAMLOverrideTakesPrecedenceOverPalette(t *testing.T) {
	unsetNoColor(t)
	const override = "#ff0000"
	thm := New("nord", Overrides{Cursor: override}, false)
	nordDefault := New("nord", Overrides{}, false)
	if thm.Cursor.GetForeground() == nordDefault.Cursor.GetForeground() {
		t.Error("YAML cursor override had no effect on style")
	}
	if thm.Cursor.GetForeground() != lipgloss.Color(override) {
		t.Errorf("cursor color: want %q, got %v", override, thm.Cursor.GetForeground())
	}
}

func TestEnvVarOverrideTakesPrecedenceOverYAML(t *testing.T) {
	unsetNoColor(t)
	const envColor = "#00ff00"
	const yamlColor = "#ff0000"
	t.Setenv("CHOOZ_THEME_SELECTED", envColor)
	thm := New("gum", Overrides{Selected: yamlColor}, false)
	// Env var must win.
	if thm.Selected.GetForeground() != lipgloss.Color(envColor) {
		t.Errorf("CHOOZ_THEME_SELECTED: want %q, got %v", envColor, thm.Selected.GetForeground())
	}
}

func TestEnvVarOverrideAllElements(t *testing.T) {
	unsetNoColor(t)
	want := map[string]lipgloss.Color{
		"CHOOZ_THEME_CURSOR":      "#111111",
		"CHOOZ_THEME_SELECTED":    "#222222",
		"CHOOZ_THEME_ITEM":        "#333333",
		"CHOOZ_THEME_HEADER":      "#444444",
		"CHOOZ_THEME_DESCRIPTION": "#555555",
		"CHOOZ_THEME_BORDER":      "#666666",
		"CHOOZ_THEME_HELP":        "#777777",
	}
	for k, v := range want {
		t.Setenv(k, string(v))
	}
	thm := New("gum", Overrides{}, false)
	cases := []struct {
		env   string
		got   lipgloss.TerminalColor
	}{
		{"CHOOZ_THEME_CURSOR", thm.Cursor.GetForeground()},
		{"CHOOZ_THEME_SELECTED", thm.Selected.GetForeground()},
		{"CHOOZ_THEME_ITEM", thm.Item.GetForeground()},
		{"CHOOZ_THEME_HEADER", thm.Header.GetForeground()},
		{"CHOOZ_THEME_DESCRIPTION", thm.Description.GetForeground()},
		{"CHOOZ_THEME_BORDER", thm.Border.GetForeground()},
		{"CHOOZ_THEME_HELP", thm.Help.GetForeground()},
	}
	for _, c := range cases {
		if c.got != want[c.env] {
			t.Errorf("%s: want %v, got %v", c.env, want[c.env], c.got)
		}
	}
}

func TestUnknownSlugFallsBackToGum(t *testing.T) {
	unsetNoColor(t)
	p := Lookup("does-not-exist")
	if p.Slug != "gum" {
		t.Errorf("expected gum fallback, got %q", p.Slug)
	}
}

func TestAllSlugsResolvable(t *testing.T) {
	unsetNoColor(t)
	for _, slug := range Slugs() {
		p := Lookup(slug)
		if p.Slug != slug {
			t.Errorf("Lookup(%q) returned slug %q", slug, p.Slug)
		}
	}
}

func TestSlugCaseInsensitiveViaNew(t *testing.T) {
	unsetNoColor(t)
	// New() lowercases before Lookup; Lookup itself is case-sensitive.
	thm := New("NORD", Overrides{}, false)
	if thm.NoColor {
		t.Error("expected styled theme for uppercase slug NORD")
	}
}

func TestKnownThemes(t *testing.T) {
	unsetNoColor(t)
	for _, slug := range []string{"nord", "dracula", "tokyo-night", "catppuccin-mocha", "everforest"} {
		thm := New(slug, Overrides{}, false)
		if thm.NoColor {
			t.Errorf("theme %q: unexpected NoColor=true", slug)
		}
	}
}

func TestVariantField(t *testing.T) {
	unsetNoColor(t)
	valid := map[string]bool{"dark": true, "light": true, "adaptive": true}
	darkSlugs := []string{"gum", "gruvbox-dark", "nord", "dracula", "catppuccin-mocha", "tokyo-night"}
	lightSlugs := []string{"gruvbox-light", "solarized-light", "catppuccin-latte", "one-light", "rose-pine-dawn"}
	adaptiveSlugs := []string{"solarized", "everforest", "rose-pine", "kanagawa", "base16"}

	for _, p := range All() {
		if !valid[p.Variant] {
			t.Errorf("slug %q: invalid Variant %q", p.Slug, p.Variant)
		}
	}
	for _, slug := range darkSlugs {
		if p := Lookup(slug); p.Variant != "dark" {
			t.Errorf("slug %q: want Variant=dark, got %q", slug, p.Variant)
		}
	}
	for _, slug := range lightSlugs {
		if p := Lookup(slug); p.Variant != "light" {
			t.Errorf("slug %q: want Variant=light, got %q", slug, p.Variant)
		}
	}
	for _, slug := range adaptiveSlugs {
		if p := Lookup(slug); p.Variant != "adaptive" {
			t.Errorf("slug %q: want Variant=adaptive, got %q", slug, p.Variant)
		}
	}
}
