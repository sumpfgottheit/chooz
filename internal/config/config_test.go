package config

import (
	"testing"
)

func TestValidParse(t *testing.T) {
	data := []byte(`
title: "Test Menu"
description: "A test menu"
items:
  - name: foo
    title: "Foo"
    description: "Foo description"
  - name: bar
    description: "Bar description"
  - name: baz
`)
	menu, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menu.Title != "Test Menu" {
		t.Errorf("title: got %q, want %q", menu.Title, "Test Menu")
	}
	if len(menu.Items) != 3 {
		t.Fatalf("items count: got %d, want 3", len(menu.Items))
	}
	if menu.Items[0].Name != "foo" || menu.Items[0].Title != "Foo" {
		t.Errorf("item[0]: got %+v", menu.Items[0])
	}
	if menu.Items[2].Name != "baz" {
		t.Errorf("item[2]: got %+v", menu.Items[2])
	}
}

func TestMissingItems(t *testing.T) {
	data := []byte(`title: "No Items"`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for missing items")
	}
}

func TestEmptyItems(t *testing.T) {
	data := []byte(`items: []`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestInvalidNameChars(t *testing.T) {
	cases := []string{
		"has-hyphen",
		"has_underscore",
		"has space",
		"has.dot",
		"",
	}
	for _, name := range cases {
		data := []byte("items:\n  - name: \"" + name + "\"\n")
		_, err := Parse(data)
		if err == nil {
			t.Errorf("expected error for name %q", name)
		}
	}
}

func TestValidNameChars(t *testing.T) {
	cases := []string{"abc", "ABC", "abc123", "A1B2C3"}
	for _, name := range cases {
		data := []byte("items:\n  - name: " + name + "\n")
		_, err := Parse(data)
		if err != nil {
			t.Errorf("unexpected error for name %q: %v", name, err)
		}
	}
}

func TestDuplicateName(t *testing.T) {
	data := []byte(`
items:
  - name: foo
  - name: bar
  - name: foo
`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestUnknownKeysIgnored(t *testing.T) {
	data := []byte(`
unknown_top_level: true
items:
  - name: foo
    unknown_item_field: bar
    another: 42
`)
	menu, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error for unknown keys: %v", err)
	}
	if len(menu.Items) != 1 || menu.Items[0].Name != "foo" {
		t.Errorf("unexpected items: %v", menu.Items)
	}
}

func TestMultilineDescription(t *testing.T) {
	data := []byte(`
items:
  - name: foo
    description: |
      Line one.
      Line two.
      Line three.
`)
	menu, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	desc := menu.Items[0].Description
	if desc == "" {
		t.Fatal("description should not be empty")
	}
	// yaml.v3 preserves newlines in literal block scalars
	if desc == "Line one." {
		t.Error("expected multi-line description to be preserved")
	}
}

func TestDisplayLabel(t *testing.T) {
	withTitle := Item{Name: "foo", Title: "Foo Label"}
	if withTitle.DisplayLabel() != "Foo Label" {
		t.Errorf("got %q, want %q", withTitle.DisplayLabel(), "Foo Label")
	}

	withoutTitle := Item{Name: "bar"}
	if withoutTitle.DisplayLabel() != "bar" {
		t.Errorf("got %q, want %q", withoutTitle.DisplayLabel(), "bar")
	}
}

func TestLoadTestdata(t *testing.T) {
	menu, err := Load("../../testdata/example.yaml")
	if err != nil {
		t.Fatalf("Load example.yaml: %v", err)
	}
	if len(menu.Items) == 0 {
		t.Error("expected non-empty items from example.yaml")
	}
}

func TestLoadInvalidName(t *testing.T) {
	_, err := Load("../../testdata/invalid_name.yaml")
	if err == nil {
		t.Fatal("expected error for invalid_name.yaml")
	}
}

func TestLoadDuplicateName(t *testing.T) {
	_, err := Load("../../testdata/duplicate_name.yaml")
	if err == nil {
		t.Fatal("expected error for duplicate_name.yaml")
	}
}
