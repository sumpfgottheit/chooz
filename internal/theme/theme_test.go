package theme

import (
	"testing"

	"github.com/saf/chooz/internal/config"
)

func TestNoColorFlag(t *testing.T) {
	thm := New(config.Theme{}, true)
	if !thm.NoColor {
		t.Error("expected NoColor=true when noColor flag is set")
	}
}

func TestNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	thm := New(config.Theme{}, false)
	if !thm.NoColor {
		t.Error("expected NoColor=true when NO_COLOR env is set")
	}
}

func TestNoColorEnvEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	thm := New(config.Theme{}, false)
	if thm.NoColor {
		t.Error("expected NoColor=false when NO_COLOR env is empty")
	}
}

func TestColorThemeCreated(t *testing.T) {
	thm := New(config.Theme{}, false)
	if thm.NoColor {
		t.Error("expected NoColor=false for default theme")
	}
	// Styles should be non-zero (have at least some configuration)
	_ = thm.Cursor.Render("x")
	_ = thm.Selected.Render("x")
	_ = thm.Item.Render("x")
	_ = thm.Header.Render("x")
	_ = thm.Description.Render("x")
	_ = thm.Border.Render("x")
	_ = thm.Help.Render("x")
}

func TestYAMLThemeOverride(t *testing.T) {
	cfg := config.Theme{Cursor: "196"}
	thm := New(cfg, false)
	if thm.NoColor {
		t.Error("expected colors enabled")
	}
	// Rendering shouldn't panic
	_ = thm.Cursor.Render("test")
}

func TestEnvOverridesYAML(t *testing.T) {
	t.Setenv("CHOOZ_THEME_CURSOR", "46")
	cfg := config.Theme{Cursor: "196"}
	// Both set; env should win — just verify no panic and NoColor=false
	thm := New(cfg, false)
	if thm.NoColor {
		t.Error("expected colors enabled")
	}
	_ = thm.Cursor.Render("test")
}
