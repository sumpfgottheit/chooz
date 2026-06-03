package theme

import (
	"testing"
)

func TestNoColorFlag(t *testing.T) {
	thm := New("gum", true)
	if !thm.NoColor {
		t.Error("expected NoColor=true when noColor flag is set")
	}
}

func TestNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	thm := New("gum", false)
	if !thm.NoColor {
		t.Error("expected NoColor=true when NO_COLOR env is set")
	}
}

func TestNoColorEnvEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	thm := New("gum", false)
	if thm.NoColor {
		t.Error("expected NoColor=false when NO_COLOR env is empty")
	}
}

func TestDefaultTheme(t *testing.T) {
	thm := New("", false)
	if thm.NoColor {
		t.Error("expected NoColor=false for default theme")
	}
	// Styles should render without panic
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

func TestUnknownSlugFallsBackToGum(t *testing.T) {
	p := Lookup("does-not-exist")
	if p.Slug != "gum" {
		t.Errorf("expected gum fallback, got %q", p.Slug)
	}
}

func TestAllSlugsResolvable(t *testing.T) {
	for _, slug := range Slugs() {
		p := Lookup(slug)
		if p.Slug != slug {
			t.Errorf("Lookup(%q) returned slug %q", slug, p.Slug)
		}
	}
}

func TestSlugCaseInsensitiveViaNew(t *testing.T) {
	// New() lowercases before Lookup; Lookup itself is case-sensitive.
	thm := New("NORD", false)
	if thm.NoColor {
		t.Error("expected styled theme for uppercase slug NORD")
	}
}

func TestKnownThemes(t *testing.T) {
	for _, slug := range []string{"nord", "dracula", "tokyo-night", "catppuccin-mocha", "everforest"} {
		thm := New(slug, false)
		if thm.NoColor {
			t.Errorf("theme %q: unexpected NoColor=true", slug)
		}
	}
}

func TestVariantField(t *testing.T) {
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
