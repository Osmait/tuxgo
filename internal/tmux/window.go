package tmux

import (
	"fmt"
	"os/exec"

	"github.com/josesaulburgos/tuxgo/internal/config"
)

// CreateWindow creates a new window in the session
func (s *Session) CreateWindow(name string) error {
	args := append(baseArgs(), "new-window", "-t", s.Name, "-n", name)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating window '%s': %v", name, err)
	}
	return nil
}

// SelectWindow selects a specific window
func (s *Session) SelectWindow(windowName string) error {
	target := fmt.Sprintf("%s:%s", s.Name, windowName)
	args := append(baseArgs(), "select-window", "-t", target)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error selecting window '%s': %v", windowName, err)
	}
	return nil
}

// CreateWindowWithPanels creates a new window and sets up its panels.
// Supports both legacy flat layouts and hierarchical mixed layouts.
func (s *Session) CreateWindowWithPanels(cfg config.WindowConfig) error {
	if err := s.CreateWindow(cfg.Name); err != nil {
		return err
	}
	return s.SetupPanels(cfg)
}

// SetupPanels configures panels on an already existing window.
// Use this for the first window (created by CreateSession) or after CreateWindow.
func (s *Session) SetupPanels(cfg config.WindowConfig) error {
	// Check if using hierarchical layout (Root field)
	if cfg.Root != nil {
		return s.createHierarchicalLayout(cfg.Name, cfg.Root)
	}

	// Legacy flat layout support
	if len(cfg.Panels) == 0 {
		return nil
	}

	// If only one panel, simple behavior
	if len(cfg.Panels) == 1 {
		return s.SendKeys(cfg.Name, cfg.Panels[0])
	}

	// Create necessary splits (n-1 splits for n panels)
	for i := 0; i < len(cfg.Panels)-1; i++ {
		if err := s.SplitWindow(cfg.Name, cfg.Layout); err != nil {
			return err
		}
	}

	// Send commands to each panel
	for i, command := range cfg.Panels {
		panelTarget := fmt.Sprintf("%s.%d", cfg.Name, i)
		if err := s.SendKeys(panelTarget, command); err != nil {
			return err
		}
	}

	return nil
}

// createHierarchicalLayout creates a hierarchical/mixed layout using a binary tree.
//
// Rules:
//   - A node with children is a CONTAINER: it defines a split direction and its
//     children occupy the resulting panes. Any "command" on a container is ignored.
//   - A node without children is a LEAF: its "command" is executed in its pane.
//   - The first child inherits the parent's pane (no split needed).
//   - The second child is created by splitting the parent's pane.
//   - Maximum 2 children per node (binary split).
func (s *Session) createHierarchicalLayout(windowName string, root *config.PanelConfig) error {
	// Map to track paneIndex -> command for leaf nodes
	commands := make(map[int]string)

	type queueItem struct {
		paneIndex int
		panel     *config.PanelConfig
	}

	// Start with root at pane 0 (already exists after CreateWindow)
	queue := []queueItem{{paneIndex: 0, panel: root}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		parentPaneIndex := item.paneIndex
		parentPanel := item.panel

		// LEAF node: no children, just record the command
		if len(parentPanel.Children) == 0 {
			if parentPanel.Command != "" {
				commands[parentPaneIndex] = parentPanel.Command
			}
			continue
		}

		// CONTAINER node: has children, split direction applies
		splitDir := parentPanel.Split
		if splitDir != "horizontal" && splitDir != "vertical" {
			splitDir = "horizontal"
		}

		// First child inherits the parent pane (no split)
		firstChild := parentPanel.Children[0]
		firstChildCopy := firstChild
		queue = append(queue, queueItem{paneIndex: parentPaneIndex, panel: &firstChildCopy})

		// Second child (if exists) is created by splitting the parent pane
		if len(parentPanel.Children) == 2 {
			secondChild := parentPanel.Children[1]

			newPaneIndex, err := s.SplitPane(windowName, parentPaneIndex, splitDir)
			if err != nil {
				return err
			}

			secondChildCopy := secondChild
			queue = append(queue, queueItem{paneIndex: newPaneIndex, panel: &secondChildCopy})
		}
	}

	// Send all commands to their respective panes
	for paneIndex, command := range commands {
		target := fmt.Sprintf("%s.%d", windowName, paneIndex)
		if err := s.SendKeys(target, command); err != nil {
			return fmt.Errorf("error sending command to pane %d: %v", paneIndex, err)
		}
	}

	return nil
}
