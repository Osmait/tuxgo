package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SendKeys sends keys/command to a specific window or pane
func (s *Session) SendKeys(target, keys string) error {
	// Target can be "window_name" or "window_name.0" for specific pane
	targetFull := fmt.Sprintf("%s:%s", s.Name, target)
	return s.SendKeysDirect(targetFull, keys)
}

// SendKeysDirect sends keys/command to an already fully-qualified tmux target
func (s *Session) SendKeysDirect(target, keys string) error {
	args := append(baseArgs(), "send-keys", "-t", target, keys, "C-m")
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error sending commands to '%s': %v", target, err)
	}
	return nil
}

// SplitWindow splits a window creating a new pane.
// layout: "horizontal" (side by side) or "vertical" (top/bottom).
func (s *Session) SplitWindow(windowName, layout string) error {
	target := fmt.Sprintf("%s:%s", s.Name, windowName)
	args := append(baseArgs(), "split-window", "-t", target)

	if layout == "horizontal" {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error splitting window '%s': %v", windowName, err)
	}
	return nil
}

// SplitPane splits a specific pane and returns the new pane ID.
// Uses stable pane IDs (%N) that don't change when other panes are split.
func (s *Session) SplitPane(windowName string, paneID string, layout string) (string, error) {
	target := fmt.Sprintf("%s:%s.%s", s.Name, windowName, paneID)
	args := append(baseArgs(), "split-window", "-t", target, "-P", "-F", "#{pane_id}")

	if layout == "horizontal" {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error splitting pane %s in window '%s': %v", paneID, windowName, err)
	}

	newPaneID := strings.TrimSpace(string(output))
	return newPaneID, nil
}

// GetPaneID returns the stable pane ID (%N) for pane index 0 in a window.
// Used to get the initial pane ID after creating a window.
func (s *Session) GetPaneID(windowName string, paneIndex int) (string, error) {
	target := fmt.Sprintf("%s:%s.%d", s.Name, windowName, paneIndex)
	args := append(baseArgs(), "display-message", "-t", target, "-p", "#{pane_id}")
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting pane ID for %s: %v", target, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListPanes returns a list of pane indices in a window
func (s *Session) ListPanes(windowName string) ([]int, error) {
	target := fmt.Sprintf("%s:%s", s.Name, windowName)
	args := append(baseArgs(), "list-panes", "-t", target, "-F", "#{pane_index}")
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
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
func (s *Session) SelectPane(windowName string, paneIndex int) error {
	target := fmt.Sprintf("%s:%s.%d", s.Name, windowName, paneIndex)
	args := append(baseArgs(), "select-pane", "-t", target)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error selecting pane %d: %v", paneIndex, err)
	}
	return nil
}
