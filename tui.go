package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionItem represents a tmux session in the list
type SessionItem struct {
	name string
}

func (i SessionItem) FilterValue() string { return i.name }
func (i SessionItem) Title() string       { return i.name }
func (i SessionItem) Description() string { return "" }

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
			i, ok := m.list.SelectedItem().(SessionItem)
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

// SessionItemDelegate customizes how session items are rendered
type SessionItemDelegate struct{}

func (d SessionItemDelegate) Height() int                             { return 1 }
func (d SessionItemDelegate) Spacing() int                            { return 0 }
func (d SessionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d SessionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SessionItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.name)

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

// SelectSessionTUI shows an interactive TUI to select a tmux session
// Returns the selected session name and a boolean indicating if a selection was made
func SelectSessionTUI(sessions []string) (string, bool, error) {
	if len(sessions) == 0 {
		return "", false, nil
	}

	// Convert sessions to list items
	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = SessionItem{name: session}
	}

	// Create list
	const defaultWidth = 40
	const listHeight = 14

	l := list.New(items, SessionItemDelegate{}, defaultWidth, listHeight)
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
