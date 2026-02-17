package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor = lipgloss.Color("#6f03fc")
	textColor    = lipgloss.Color("#FFFDF5")
	dimmedColor  = lipgloss.Color("#666666")
	accentColor  = lipgloss.Color("#9D79BC")

	titleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				PaddingLeft(2)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(2)

	statsStyle = lipgloss.NewStyle().
			Foreground(dimmedColor).
			Italic(true).
			PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimmedColor).
			PaddingLeft(2).
			PaddingBottom(1)

	paginationStyle = lipgloss.NewStyle().
			Foreground(dimmedColor).
			PaddingLeft(2)
)

type DirItem struct {
	Path     string
	Name     string
	UseCount int
	LastUsed time.Time
}

func (i DirItem) FilterValue() string { return i.Name + " " + i.Path }

func (i DirItem) Title() string {
	homeDir, _ := os.UserHomeDir()
	relPath, err := filepath.Rel(homeDir, i.Path)
	if err != nil {
		return i.Path
	}
	return filepath.Join("~", relPath)
}

func (i DirItem) Description() string {
	return fmt.Sprintf("used %d times • %s", i.UseCount, formatRelativeTime(i.LastUsed))
}

func (i DirItem) Render(selected bool, width int) string {
	homeDir, _ := os.UserHomeDir()
	relPath, err := filepath.Rel(homeDir, i.Path)
	if err != nil {
		relPath = i.Path
	}
	displayPath := filepath.Join("~", relPath)

	if width > 0 && len(displayPath) > width-4 {
		displayPath = truncatePath(displayPath, width-4)
	}

	stats := fmt.Sprintf("used %d times • %s", i.UseCount, formatRelativeTime(i.LastUsed))

	var pathLine string
	if selected {
		pathLine = selectedItemStyle.Render("❯ " + displayPath)
	} else {
		pathLine = itemStyle.Render("  " + displayPath)
	}

	return pathLine + "\n" + statsStyle.Render("│ "+stats)
}

func (i DirItem) Height() int { return 2 }

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}

	result := "…/" + parts[len(parts)-1]
	for i := len(parts) - 2; i >= 0; i-- {
		newResult := parts[i] + "/" + result
		if len(newResult) > maxLen {
			break
		}
		result = newResult
	}

	return result
}
