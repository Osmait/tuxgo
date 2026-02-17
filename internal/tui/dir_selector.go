package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type dirItemDelegate struct{}

func (d dirItemDelegate) Height() int  { return 2 }
func (d dirItemDelegate) Spacing() int { return 1 }
func (d dirItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d dirItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(DirItem)
	if !ok {
		return
	}

	fmt.Fprint(w, i.Render(index == m.Index(), m.Width()))
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

		case "g":
			m.list.Select(0)
			return m, nil

		case "G":
			m.list.Select(len(m.list.Items()) - 1)
			return m, nil

		case "enter":
			i, ok := m.list.SelectedItem().(DirItem)
			if ok {
				m.choice = i.Path
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
		relPath := formatPath(m.choice)
		return fmt.Sprintf("Navigating to '%s'...\n", relPath)
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k up • ↓/j down • g top • G bottom • / filter • enter select • q quit"))
	return b.String()
}

func formatPath(path string) string {
	homeDir, _ := os.UserHomeDir()
	if strings.HasPrefix(path, homeDir) {
		return filepath.Join("~", strings.TrimPrefix(path, homeDir))
	}
	return path
}

func SelectDirectory(items []DirItem) (string, bool, error) {
	if len(items) == 0 {
		return "", false, nil
	}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	const defaultWidth = 70
	const listHeight = 16

	l := list.New(listItems, dirItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a directory"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := dirSelectorModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
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
