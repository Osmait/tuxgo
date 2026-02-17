package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type SessionItem struct {
	Name string
}

func (i SessionItem) FilterValue() string { return i.Name }
func (i SessionItem) Title() string       { return i.Name }
func (i SessionItem) Description() string { return "" }

type sessionItemDelegate struct{}

func (d sessionItemDelegate) Height() int                             { return 1 }
func (d sessionItemDelegate) Spacing() int                            { return 0 }
func (d sessionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d sessionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SessionItem)
	if !ok {
		return
	}

	str := i.Name

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("❯ " + s[0])
		}
	}

	fmt.Fprint(w, fn(str))
}

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
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "j", "down":
			m.list.CursorDown()
			return m, nil

		case "k", "up":
			m.list.CursorUp()
			return m, nil

		case "enter":
			i, ok := m.list.SelectedItem().(SessionItem)
			if ok {
				m.choice = i.Name
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

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k up • ↓/j down • enter select • q quit"))
	return b.String()
}

func SelectSession(sessions []string) (string, bool, error) {
	if len(sessions) == 0 {
		return "", false, nil
	}

	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = SessionItem{Name: session}
	}

	const defaultWidth = 50
	const listHeight = 14

	l := list.New(items, sessionItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a tmux session"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := sessionSelectorModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
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
