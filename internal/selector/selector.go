package selector

import (
	"fmt"
	"slices"

	"github.com/saf/chooz/internal/config"
)

// ErrNotFound is returned when a name does not match any item.
type ErrNotFound struct {
	Name string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("name %q not found", e.Name)
}

// Resolve finds the item with the given name.
func Resolve(items []config.Item, name string) (*config.Item, error) {
	if i := slices.IndexFunc(items, func(it config.Item) bool { return it.Name == name }); i >= 0 {
		return &items[i], nil
	}
	return nil, ErrNotFound{Name: name}
}

// List returns all item names in YAML order.
func List(items []config.Item) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}

// DefaultIndex returns the index of the item with the given name.
// Returns 0 and no error when name is empty (first item is default).
func DefaultIndex(items []config.Item, name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	if i := slices.IndexFunc(items, func(it config.Item) bool { return it.Name == name }); i >= 0 {
		return i, nil
	}
	return 0, ErrNotFound{Name: name}
}
