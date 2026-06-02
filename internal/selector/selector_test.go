package selector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/saf/chooz/internal/config"
)

var testItems = []config.Item{
	{Name: "prod", Title: "Production"},
	{Name: "stage", Title: "Staging"},
	{Name: "dev"},
}

func TestResolveHit(t *testing.T) {
	item, err := Resolve(testItems, "stage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "stage" {
		t.Errorf("got %q, want %q", item.Name, "stage")
	}
}

func TestResolveMiss(t *testing.T) {
	_, err := Resolve(testItems, "nope")
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	var notFound ErrNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
	if notFound.Name != "nope" {
		t.Errorf("ErrNotFound.Name: got %q, want %q", notFound.Name, "nope")
	}
}

func TestList(t *testing.T) {
	names := List(testItems)
	want := []string{"prod", "stage", "dev"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("got %v, want %v", names, want)
	}
}

func TestDefaultIndexFound(t *testing.T) {
	idx, err := DefaultIndex(testItems, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 2 {
		t.Errorf("got %d, want 2", idx)
	}
}

func TestDefaultIndexEmpty(t *testing.T) {
	idx, err := DefaultIndex(testItems, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 0 {
		t.Errorf("got %d, want 0", idx)
	}
}

func TestDefaultIndexMiss(t *testing.T) {
	_, err := DefaultIndex(testItems, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown default")
	}
	var notFound ErrNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}
