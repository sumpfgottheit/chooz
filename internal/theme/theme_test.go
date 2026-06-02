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

func TestSlugCaseInsensitive(t *testing.T) {
	p := Lookup("GUM")
	if p.Slug != "gum" {
		t.Errorf("expected gum for uppercase GUM, got %q", p.Slug)
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
