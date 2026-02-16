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
	args := append(baseArgs(), "send-keys", "-t", targetFull, keys, "C-m")
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

// SplitPane splits a specific pane and returns the new pane index
func (s *Session) SplitPane(windowName string, paneIndex int, layout string) (int, error) {
	// Get list of panes before split
	panesBefore, err := s.ListPanes(windowName)
	if err != nil {
		return -1, err
	}

	target := fmt.Sprintf("%s:%s.%d", s.Name, windowName, paneIndex)
	args := append(baseArgs(), "split-window", "-t", target)

	if layout == "horizontal" {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return -1, fmt.Errorf("error splitting pane %d in window '%s': %v", paneIndex, windowName, err)
	}

	// Get list of panes after split and find the new one
	panesAfter, err := s.ListPanes(windowName)
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
