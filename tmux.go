package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// TmuxSession handles tmux session creation and management
type TmuxSession struct {
	Name    string
	WorkDir string
}

// TmuxBaseArgs returns base arguments for all tmux commands (includes -u for UTF-8)
func TmuxBaseArgs() []string {
	return []string{"-u"}
}

// HasSession checks if a tmux session already exists
func (t *TmuxSession) HasSession() bool {
	args := append(TmuxBaseArgs(), "has-session", "-t", t.Name)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	err := cmd.Run()
	return err == nil
}

// CreateSession creates a new tmux session with the first window
func (t *TmuxSession) CreateSession(firstWindowName, firstCommand string) error {
	args := append(TmuxBaseArgs(), "new-session", "-d", "-s", t.Name, "-n", firstWindowName)

	// If there's a command for the first window, execute it after creating the session
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating tmux session: %v", err)
	}

	// If there's a command, execute it in the window
	if firstCommand != "" {
		if err := t.SendKeys(firstWindowName, firstCommand); err != nil {
			return fmt.Errorf("error executing command in first window: %v", err)
		}
	}

	return nil
}

// CreateWindow creates a new window in the session
func (t *TmuxSession) CreateWindow(name string) error {
	args := append(TmuxBaseArgs(), "new-window", "-t", t.Name, "-n", name)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating window '%s': %v", name, err)
	}
	return nil
}

// SendKeys sends keys/command to a specific window or panel
func (t *TmuxSession) SendKeys(target, keys string) error {
	// Target can be "window_name" or "window_name.0" for specific panel
	targetFull := fmt.Sprintf("%s:%s", t.Name, target)
	args := append(TmuxBaseArgs(), "send-keys", "-t", targetFull, keys, "C-m")
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error sending commands to '%s': %v", target, err)
	}
	return nil
}

// SplitWindow splits a window creating a new panel
// layout: "horizontal" or "vertical"
func (t *TmuxSession) SplitWindow(windowName, layout string) error {
	target := fmt.Sprintf("%s:%s", t.Name, windowName)
	args := append(TmuxBaseArgs(), "split-window", "-t", target)

	// -h for horizontal (side by side panels)
	// -v for vertical (top/bottom panels)
	if layout == "horizontal" {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error splitting window '%s': %v", windowName, err)
	}
	return nil
}

// SplitPane splits a specific pane and returns the new pane index
func (t *TmuxSession) SplitPane(windowName string, paneIndex int, layout string) (int, error) {
	// Get list of panes before split
	panesBefore, err := t.ListPanes(windowName)
	if err != nil {
		return -1, err
	}

	target := fmt.Sprintf("%s:%s.%d", t.Name, windowName, paneIndex)
	args := append(TmuxBaseArgs(), "split-window", "-t", target)

	if layout == "horizontal" {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return -1, fmt.Errorf("error splitting pane %d in window '%s': %v", paneIndex, windowName, err)
	}

	// Get list of panes after split and find the new one
	panesAfter, err := t.ListPanes(windowName)
	if err != nil {
		return -1, err
	}

	// Find the new pane (the one that wasn't in panesBefore)
	for _, pane := range panesAfter {
		found := false
		for _, beforePane := range panesBefore {
			if pane == beforePane {
				found = true
				break
			}
		}
		if !found {
			return pane, nil
		}
	}

	return -1, fmt.Errorf("could not determine new pane index")
}

// ListPanes returns a list of pane indices in a window
func (t *TmuxSession) ListPanes(windowName string) ([]int, error) {
	target := fmt.Sprintf("%s:%s", t.Name, windowName)
	args := append(TmuxBaseArgs(), "list-panes", "-t", target, "-F", "#{pane_index}")
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error listing panes: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var panes []int
	for _, line := range lines {
		if line == "" {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		panes = append(panes, idx)
	}

	return panes, nil
}

// SelectPane selects a specific pane by index
func (t *TmuxSession) SelectPane(windowName string, paneIndex int) error {
	target := fmt.Sprintf("%s:%s.%d", t.Name, windowName, paneIndex)
	args := append(TmuxBaseArgs(), "select-pane", "-t", target)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error selecting pane %d: %v", paneIndex, err)
	}
	return nil
}

// SelectWindow selects a specific window
func (t *TmuxSession) SelectWindow(windowName string) error {
	target := fmt.Sprintf("%s:%s", t.Name, windowName)
	args := append(TmuxBaseArgs(), "select-window", "-t", target)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = t.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error selecting window '%s': %v", windowName, err)
	}
	return nil
}

// IsInsideTmux detects if we're running the command inside a tmux session
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// AttachSession attaches to the tmux session
// If already inside tmux, uses switch-client instead of attach-session
func (t *TmuxSession) AttachSession() error {
	if IsInsideTmux() {
		// Already in tmux, use switch-client to change session
		args := append(TmuxBaseArgs(), "switch-client", "-t", t.Name)
		cmd := exec.Command("tmux", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = t.WorkDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error switching to tmux session: %v", err)
		}
	} else {
		// Not in tmux, do normal attach
		args := append(TmuxBaseArgs(), "attach-session", "-t", t.Name)
		cmd := exec.Command("tmux", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = t.WorkDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error attaching to tmux session: %v", err)
		}
	}
	return nil
}

// CreateWindowWithPanels creates a window with multiple panels and executes commands
// Supports both legacy flat layouts and new hierarchical mixed layouts
func (t *TmuxSession) CreateWindowWithPanels(config WindowConfig) error {
	// Create the window
	if err := t.CreateWindow(config.Name); err != nil {
		return err
	}

	// Check if using new hierarchical layout (Root field)
	if config.Root != nil {
		return t.createHierarchicalLayout(config.Name, config.Root)
	}

	// Legacy flat layout support
	if len(config.Panels) == 0 {
		return nil
	}

	// If only one panel, simple behavior
	if len(config.Panels) == 1 {
		return t.SendKeys(config.Name, config.Panels[0])
	}

	// Create necessary splits (n-1 splits for n panels)
	for i := 0; i < len(config.Panels)-1; i++ {
		if err := t.SplitWindow(config.Name, config.Layout); err != nil {
			return err
		}
	}

	// Send commands to each panel
	for i, command := range config.Panels {
		panelTarget := fmt.Sprintf("%s.%d", config.Name, i)
		if err := t.SendKeys(panelTarget, command); err != nil {
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
func (t *TmuxSession) createHierarchicalLayout(windowName string, root *PanelConfig) error {
	// Map to track paneIndex -> command for leaf nodes
	commands := make(map[int]string)

	type queueItem struct {
		paneIndex int
		panel     *PanelConfig
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

			newPaneIndex, err := t.SplitPane(windowName, parentPaneIndex, splitDir)
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
		if err := t.SendKeys(target, command); err != nil {
			return fmt.Errorf("error sending command to pane %d: %v", paneIndex, err)
		}
	}

	return nil
}

// GetSessionName gets the session name based on the current directory
func GetSessionName(workDir string) string {
	sessionName := filepath.Base(workDir)
	// Clean name (replace dots with underscores for tmux)
	sessionName = strings.ReplaceAll(sessionName, ".", "_")
	return sessionName
}

// ListSessions returns a list of all active tmux sessions
func ListSessions() ([]string, error) {
	args := append(TmuxBaseArgs(), "list-sessions", "-F", "#{session_name}")
	cmd := exec.Command("tmux", args...)
	output, err := cmd.Output()
	if err != nil {
		// If there's an error (no sessions), return empty slice
		return []string{}, nil
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}

	return sessions, nil
}

// AttachToSession attaches to an existing session by name
func AttachToSession(sessionName string) error {
	session := &TmuxSession{Name: sessionName}
	return session.AttachSession()
}

// ValidateConfig validates that the configuration has valid windows
func ValidateConfig(config *ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if len(config.Windows) == 0 {
		return fmt.Errorf("configuration has no windows defined")
	}

	for _, window := range config.Windows {
		if window.Name == "" {
			return fmt.Errorf("window without name defined")
		}

		// Count configuration types
		hasCommand := window.Command != ""
		hasPanels := len(window.Panels) > 0
		hasRoot := window.Root != nil

		configCount := 0
		if hasCommand {
			configCount++
		}
		if hasPanels {
			configCount++
		}
		if hasRoot {
			configCount++
		}

		if configCount == 0 {
			return fmt.Errorf("window '%s' has no configuration (need 'command', 'panels', or 'root')", window.Name)
		}

		if configCount > 1 {
			log.Printf("Warning: window '%s' has multiple configurations, using 'root' > 'panels' > 'command'", window.Name)
		}

		// Validate legacy layout if using flat panels
		if hasPanels && !hasRoot {
			if len(window.Panels) > 1 && window.Layout != "horizontal" && window.Layout != "vertical" {
				return fmt.Errorf("window '%s' has multiple panels but invalid layout: '%s' (use 'horizontal' or 'vertical')", window.Name, window.Layout)
			}
		}

		// Validate hierarchical root structure
		if hasRoot {
			if err := validatePanelConfig(window.Root, window.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// validatePanelConfig recursively validates a panel configuration
func validatePanelConfig(panel *PanelConfig, windowName string) error {
	if panel == nil {
		return nil
	}

	// Validate split direction if children exist
	if len(panel.Children) > 0 {
		if panel.Split != "" && panel.Split != "horizontal" && panel.Split != "vertical" {
			return fmt.Errorf("invalid split direction '%s' in window '%s' (use 'horizontal' or 'vertical')", panel.Split, windowName)
		}

		if len(panel.Children) > 2 {
			return fmt.Errorf("panel in window '%s' has more than 2 children (maximum 2 for binary splits)", windowName)
		}

		// Warn if a container node also has a command (it will be ignored)
		if panel.Command != "" {
			log.Printf("Warning: panel in window '%s' has both 'command' and 'children'; command will be ignored (move it to a child)", windowName)
		}

		// Recursively validate children
		for i := range panel.Children {
			if err := validatePanelConfig(&panel.Children[i], windowName); err != nil {
				return err
			}
		}
	}

	return nil
}
