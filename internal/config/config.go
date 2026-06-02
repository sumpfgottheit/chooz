package config

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type Item struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func (i Item) DisplayLabel() string {
	if i.Title != "" {
		return i.Title
	}
	return i.Name
}

type Theme struct {
	Cursor      string `yaml:"cursor"`
	Selected    string `yaml:"selected"`
	Item        string `yaml:"item"`
	Header      string `yaml:"header"`
	Description string `yaml:"description"`
	Border      string `yaml:"border"`
	Help        string `yaml:"help"`
}

type Menu struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Items       []Item `yaml:"items"`
	Theme       Theme  `yaml:"theme"`
}

func Load(path string) (*Menu, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}
	}
	return Parse(data)
}

func Parse(data []byte) (*Menu, error) {
	var menu Menu
	if err := yaml.Unmarshal(data, &menu); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if len(menu.Items) == 0 {
		return nil, fmt.Errorf("items: must be non-empty")
	}
	seen := make(map[string]bool, len(menu.Items))
	for i, item := range menu.Items {
		if item.Name == "" {
			return nil, fmt.Errorf("items[%d]: name is required", i)
		}
		if !nameRe.MatchString(item.Name) {
			return nil, fmt.Errorf("items[%d]: name %q must match ^[A-Za-z0-9]+$", i, item.Name)
		}
		if seen[item.Name] {
			return nil, fmt.Errorf("items[%d]: duplicate name %q", i, item.Name)
		}
		seen[item.Name] = true
	}
	return &menu, nil
}
