package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/saf/chooz/internal/theme"
)

type showroomModel struct {
	palettes []theme.Palette
	idx      int
}

func (m showroomModel) Init() tea.Cmd { return nil }

func (m showroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l", "n", "j":
			m.idx = (m.idx + 1) % len(m.palettes)
		case "shift+tab", "left", "h", "p", "k":
			m.idx = (m.idx - 1 + len(m.palettes)) % len(m.palettes)
		}
	}
	return m, nil
}

func (m showroomModel) View() string {
	p := m.palettes[m.idx]

	cursorStyle  := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	itemStyle    := lipgloss.NewStyle().Foreground(p.Fg)
	headerStyle  := lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	descStyle    := lipgloss.NewStyle().Foreground(p.Fg)
	borderStyle  := lipgloss.NewStyle().Foreground(p.Muted)
	helpStyle    := lipgloss.NewStyle().Foreground(p.Muted)

	// ── mock chooz UI ────────────────────────────────────────────────────
	menuTitle := headerStyle.Render("Select a target environment")

	listCol := strings.Join([]string{
		cursorStyle.Render(">") + " " + selectedStyle.Render("Production"),
		itemStyle.Render("  Staging"),
		itemStyle.Render("  Development"),
		itemStyle.Render("  Preview"),
	}, "\n")

	previewSep := borderStyle.Render(strings.Repeat("─", 30))
	previewCol := strings.Join([]string{
		selectedStyle.Render("Production"),
		previewSep,
		descStyle.Render("Live cluster. Handle with care."),
		descStyle.Render("All changes go live immediately."),
		descStyle.Render("Requires two approvals."),
	}, "\n")

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		listCol,
		borderStyle.Render(" │ "),
		previewCol,
	)

	choozHelp := helpStyle.Render(
		"↑/↓/j/k: navigate • enter: select • q/esc: cancel • pgup/pgdn: scroll",
	)

	mockUI := strings.Join([]string{menuTitle, "", columns, "", choozHelp}, "\n")

	// ── showroom chrome ──────────────────────────────────────────────────
	nameLabel  := headerStyle.Render(p.Name)
	counter    := helpStyle.Render(fmt.Sprintf("%d/%d", m.idx+1, len(m.palettes)))
	titleBar   := lipgloss.JoinHorizontal(lipgloss.Bottom, nameLabel, "   ", counter)

	body := strings.Join([]string{titleBar, "", mockUI}, "\n")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Primary).
		Padding(1, 2).
		Render(body)

	navHint := helpStyle.Render("tab / → : next theme   shift+tab / ← : prev   q : quit")

	return lipgloss.NewStyle().Padding(1, 2).Render(panel) +
		"\n" +
		lipgloss.NewStyle().PaddingLeft(4).Render(navHint) +
		"\n"
}

// RunShowroom starts the interactive theme showroom.
func RunShowroom() error {
	m := showroomModel{palettes: theme.All()}

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer func() { _ = f.Close() }()
		opts = append(opts, tea.WithOutput(f), tea.WithInput(f))
	} else {
		opts = append(opts, tea.WithOutput(os.Stderr))
	}

	_, err := tea.NewProgram(m, opts...).Run()
	return err
}
