package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sessionItem represents a tmux session in the list
type sessionItem struct {
	name string
}

func (i sessionItem) FilterValue() string { return i.name }
func (i sessionItem) Title() string       { return i.name }
func (i sessionItem) Description() string { return "" }

// sessionSelectorModel is the Bubble Tea model for session selection
type sessionSelectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m sessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m sessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(sessionItem)
			if ok {
				m.choice = i.name
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m sessionSelectorModel) View() string {
	if m.quitting {
		return ""
	}
	if m.choice != "" {
		return fmt.Sprintf("Attaching to session '%s'...\n", m.choice)
	}
	return "\n" + m.list.View()
}

// sessionItemDelegate customizes how session items are rendered
type sessionItemDelegate struct{}

func (d sessionItemDelegate) Height() int                             { return 1 }
func (d sessionItemDelegate) Spacing() int                            { return 0 }
func (d sessionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d sessionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sessionItem)
	if !ok {
		return
	}

	str := i.name

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + s[0])
		}
	}

	fmt.Fprint(w, fn(str))
}

// Styles for the TUI
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#6f03fc"))
)

type dirItem struct {
	path string
}

func (i dirItem) FilterValue() string { return i.path }
func (i dirItem) Title() string {
	relPath, err := filepath.Rel(os.Getenv("HOME"), i.path)
	if err != nil {
		return i.path
	}
	return relPath
}
func (i dirItem) Description() string { return i.path }

type dirItemDelegate struct{}

func (d dirItemDelegate) Height() int                             { return 1 }
func (d dirItemDelegate) Spacing() int                            { return 0 }
func (d dirItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d dirItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(dirItem)
	if !ok {
		return
	}

	relPath, err := filepath.Rel(os.Getenv("HOME"), i.path)
	if err != nil {
		relPath = i.path
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + s[0])
		}
	}

	fmt.Fprint(w, fn(relPath))
}

type dirSelectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m dirSelectorModel) Init() tea.Cmd {
	return nil
}

func (m dirSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(dirItem)
			if ok {
				m.choice = i.path
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m dirSelectorModel) View() string {
	if m.quitting {
		return ""
	}
	if m.choice != "" {
		relPath, _ := filepath.Rel(os.Getenv("HOME"), m.choice)
		return fmt.Sprintf("Navigating to '%s'...\n", relPath)
	}
	return "\n" + m.list.View()
}

func SelectDirectory(paths []string) (string, bool, error) {
	if len(paths) == 0 {
		return "", false, nil
	}

	items := make([]list.Item, len(paths))
	for i, path := range paths {
		items[i] = dirItem{path: path}
	}

	const defaultWidth = 60
	const listHeight = 14

	l := list.New(items, dirItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a directory"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6f03fc")).
		Padding(0, 1)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	m := dirSelectorModel{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}

	model := finalModel.(dirSelectorModel)
	if model.quitting || model.choice == "" {
		return "", false, nil
	}

	return model.choice, true, nil
}

func SelectSession(sessions []string) (string, bool, error) {
	if len(sessions) == 0 {
		return "", false, nil
	}

	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = sessionItem{name: session}
	}

	const defaultWidth = 40
	const listHeight = 14

	l := list.New(items, sessionItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a tmux session to attach"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6f03fc")).
		Padding(0, 1)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	m := sessionSelectorModel{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}

	model := finalModel.(sessionSelectorModel)
	if model.quitting || model.choice == "" {
		return "", false, nil
	}

	return model.choice, true, nil
}
