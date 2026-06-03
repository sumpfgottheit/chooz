package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/saf/chooz/internal/config"
	"github.com/saf/chooz/internal/theme"
)

const wideThreshold = 80

// menuItem wraps config.Item to satisfy list.Item.
type menuItem struct {
	item config.Item
}

func (m menuItem) FilterValue() string { return m.item.Name }

// itemDelegate renders each list row using the current theme.
type itemDelegate struct {
	thm *theme.Theme
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	mi, ok := listItem.(menuItem)
	if !ok {
		return
	}
	const prefixWidth = 2 // "> " or "  "
	label := truncate(mi.item.DisplayLabel(), m.Width()-prefixWidth)
	if d.thm.NoColor {
		if index == m.Index() {
			fmt.Fprint(w, "> "+label)
		} else {
			fmt.Fprint(w, "  "+label)
		}
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, d.thm.Cursor.Render("> ")+d.thm.Selected.Render(label))
	} else {
		fmt.Fprint(w, d.thm.Item.Render("  "+label))
	}
}

// Model is the Bubble Tea model for the interactive menu.
type Model struct {
	list       list.Model
	viewport   viewport.Model
	allItems   []list.Item // full unfiltered item set
	items      []config.Item
	menu       *config.Menu
	thm        *theme.Theme
	filter     string
	maxHeight  int
	width      int
	height     int
	selected   string
	cancelled  bool
	ready      bool
	hasAnyDesc bool // true if at least one item has a description
}

func newModel(menu *config.Menu, thm *theme.Theme, defaultName string, maxHeight int) Model {
	allItems := make([]list.Item, len(menu.Items))
	for i, item := range menu.Items {
		allItems[i] = menuItem{item: item}
	}

	l := list.New(allItems, itemDelegate{thm: thm}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.SetShowHelp(false)

	if defaultName != "" {
		for i, item := range menu.Items {
			if item.Name == defaultName {
				l.Select(i)
				break
			}
		}
	}

	hasAnyDesc := false
	for _, item := range menu.Items {
		if item.Description != "" {
			hasAnyDesc = true
			break
		}
	}

	return Model{
		list:       l,
		viewport:   viewport.New(0, 0),
		allItems:   allItems,
		items:      menu.Items,
		menu:       menu,
		thm:        thm,
		maxHeight:  maxHeight,
		hasAnyDesc: hasAnyDesc,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.withLayout()
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		// Always-handled keys regardless of filter state.
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			if m.filter != "" {
				m.filter = ""
				m = m.applyFilter()
				return m, nil
			}
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(menuItem); ok {
				m.selected = item.item.Name
			}
			return m, tea.Quit
		case "backspace":
			if len(m.filter) > 0 {
				runes := []rune(m.filter)
				m.filter = string(runes[:len(runes)-1])
				m = m.applyFilter()
			}
			return m, nil
		case "pgup", "ctrl+b", "pgdown", "ctrl+f":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		// When filter is active: runes extend the filter; other keys (arrows)
		// fall through to list navigation below.
		if m.filter != "" {
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m = m.applyFilter()
				return m, nil
			}
		} else {
			// Filter is empty: q quits; runes start the filter.
			switch msg.String() {
			case "q":
				m.cancelled = true
				return m, tea.Quit
			}
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m = m.applyFilter()
				return m, nil
			}
		}
	}

	// Default: pass through to list (handles ↑/↓/j/k/g/G when filter is empty,
	// and ↑/↓ arrow keys when filter is active).
	prevIdx := m.list.Index()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIdx {
		m = m.withPreview()
	}
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading…\n"
	}

	var sb strings.Builder

	if m.menu.Title != "" {
		if m.thm.NoColor {
			sb.WriteString(m.menu.Title + "\n")
		} else {
			sb.WriteString(m.thm.Header.Render(m.menu.Title) + "\n")
		}
	}
	if m.menu.Description != "" {
		if m.thm.NoColor {
			sb.WriteString(m.menu.Description + "\n")
		} else {
			sb.WriteString(m.thm.Description.Render(m.menu.Description) + "\n")
		}
	}

	listView := m.list.View()

	if !m.hasAnyDesc {
		sb.WriteString(listView)
	} else {
		previewView := m.viewport.View()
		if m.width >= wideThreshold {
			sep := " │ "
			if !m.thm.NoColor {
				sep = m.thm.Border.Render(sep)
			}
			sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, sep, previewView))
		} else {
			sb.WriteString(listView)
			sb.WriteString("\n")
			divider := strings.Repeat("─", m.width)
			if !m.thm.NoColor {
				divider = m.thm.Border.Render(divider)
			}
			sb.WriteString(divider + "\n")
			sb.WriteString(previewView)
		}
	}

	sb.WriteString("\n")

	// Filter line.
	if m.filter == "" {
		placeholder := "  / type to filter"
		if m.thm.NoColor {
			sb.WriteString(placeholder + "\n")
		} else {
			sb.WriteString(m.thm.Help.Render(placeholder) + "\n")
		}
	} else {
		if m.thm.NoColor {
			sb.WriteString("  / " + m.filter + "\n")
		} else {
			sb.WriteString(m.thm.Help.Render("  / ") + m.thm.Selected.Render(m.filter) + "\n")
		}
	}

	// Help line.
	hint := "↑/↓/j/k: navigate • enter: select • esc/q: cancel • backspace: clear filter"
	if m.hasAnyDesc {
		hint += " • pgup/pgdn: scroll preview"
	}
	if m.thm.NoColor {
		sb.WriteString(hint)
	} else {
		sb.WriteString(m.thm.Help.Render(hint))
	}

	return sb.String()
}

// applyFilter re-filters the list based on m.filter and resets selection to 0.
func (m Model) applyFilter() Model {
	q := strings.ToLower(m.filter)
	if q == "" {
		m.list.SetItems(m.allItems)
	} else {
		var filtered []list.Item
		for _, item := range m.allItems {
			mi := item.(menuItem)
			label := strings.ToLower(mi.item.DisplayLabel())
			name := strings.ToLower(mi.item.Name)
			if strings.Contains(label, q) || strings.Contains(name, q) {
				filtered = append(filtered, item)
			}
		}
		m.list.SetItems(filtered)
	}
	m.list.Select(0)
	return m.withPreview()
}

// withLayout recalculates list and viewport sizes from the current terminal dimensions.
func (m Model) withLayout() Model {
	headerLines := 0
	if m.menu.Title != "" {
		headerLines++
	}
	if m.menu.Description != "" {
		headerLines += strings.Count(m.menu.Description, "\n") + 1
	}

	const footerLines = 2 // filter line + help line
	const padding = 2

	contentH := m.height - headerLines - footerLines - padding
	if contentH < 3 {
		contentH = 3
	}

	listH := contentH
	if m.maxHeight > 0 && m.maxHeight < listH {
		listH = m.maxHeight
	}

	if !m.hasAnyDesc {
		m.list.SetSize(m.width, listH)
		return m
	}

	if m.width >= wideThreshold {
		listW := m.width * 2 / 5
		if listW < 15 {
			listW = 15
		}
		previewW := m.width - listW - 3
		if previewW < 10 {
			previewW = 10
		}
		m.list.SetSize(listW, listH)
		m.viewport.Width = previewW
		m.viewport.Height = contentH
	} else {
		listH2 := listH / 2
		if listH2 < 3 {
			listH2 = 3
		}
		previewH := contentH - listH2 - 1
		if previewH < 2 {
			previewH = 2
		}
		m.list.SetSize(m.width, listH2)
		m.viewport.Width = m.width
		m.viewport.Height = previewH
	}

	return m.withPreview()
}

// withPreview updates the viewport content for the currently selected item.
func (m Model) withPreview() Model {
	if m.viewport.Width <= 0 {
		return m
	}
	item, ok := m.list.SelectedItem().(menuItem)
	if !ok {
		return m
	}

	title := item.item.DisplayLabel()
	desc := strings.TrimRight(item.item.Description, "\n")

	var content string
	if desc == "" {
		content = wrapText(title, m.viewport.Width)
	} else {
		sep := strings.Repeat("─", m.viewport.Width)
		content = wrapText(title, m.viewport.Width) + "\n" + sep + "\n" + wrapText(desc, m.viewport.Width)
	}

	if m.thm.NoColor {
		m.viewport.SetContent(content)
	} else {
		titlePart := m.thm.Selected.Render(wrapText(title, m.viewport.Width))
		if desc == "" {
			m.viewport.SetContent(titlePart)
		} else {
			sep := m.thm.Border.Render(strings.Repeat("─", m.viewport.Width))
			descPart := m.thm.Description.Render(wrapText(desc, m.viewport.Width))
			m.viewport.SetContent(titlePart + "\n" + sep + "\n" + descPart)
		}
	}
	m.viewport.GotoTop()
	return m
}

// wrapText wraps text at the given column width, preserving existing newlines.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapLine(line, width))
	}
	return strings.Join(wrapped, "\n")
}

func wrapLine(line string, width int) string {
	if runewidth.StringWidth(line) <= width {
		return line
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return ""
	}
	var result []string
	var cur strings.Builder
	lineLen := 0
	for _, word := range words {
		for runewidth.StringWidth(word) > width {
			if lineLen > 0 {
				result = append(result, cur.String())
				cur.Reset()
				lineLen = 0
			}
			chunk := runewidth.Truncate(word, width, "")
			if chunk == "" {
				break // guard: a single character is wider than the viewport
			}
			result = append(result, chunk)
			word = word[len(chunk):]
		}
		if word == "" {
			continue
		}
		wl := runewidth.StringWidth(word)
		if lineLen == 0 {
			cur.WriteString(word)
			lineLen = wl
		} else if lineLen+1+wl > width {
			result = append(result, cur.String())
			cur.Reset()
			cur.WriteString(word)
			lineLen = wl
		} else {
			cur.WriteByte(' ')
			cur.WriteString(word)
			lineLen += 1 + wl
		}
	}
	if cur.Len() > 0 {
		result = append(result, cur.String())
	}
	return strings.Join(result, "\n")
}

// truncate shortens s to max terminal display cells, appending "…" if trimmed.
func truncate(s string, max int) string {
	if max <= 0 || runewidth.StringWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return runewidth.Truncate(s, max-1, "") + "…"
}

// Run starts the interactive TUI and returns the selected item name.
// Returns "" if the user cancelled (exit 130 should be used by the caller).
func Run(menu *config.Menu, thm *theme.Theme, defaultName string, maxHeight int) (string, error) {
	m := newModel(menu, thm, defaultName, maxHeight)

	opts := []tea.ProgramOption{tea.WithAltScreen()}

	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer func() { _ = f.Close() }()
		opts = append(opts, tea.WithOutput(f), tea.WithInput(f))
	} else {
		opts = append(opts, tea.WithOutput(os.Stderr))
	}

	p := tea.NewProgram(m, opts...)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("TUI: %w", err)
	}

	fm, ok := finalModel.(Model)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}
	if fm.cancelled {
		return "", nil
	}
	return fm.selected, nil
}
