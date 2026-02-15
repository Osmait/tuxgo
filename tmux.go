package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

// SendKeys sends keys/command to a specific window
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
func (t *TmuxSession) CreateWindowWithPanels(config WindowConfig) error {
	// Create the window
	if err := t.CreateWindow(config.Name); err != nil {
		return err
	}

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

// GetSessionName gets the session name based on the current directory
func GetSessionName(workDir string) string {
	sessionName := filepath.Base(workDir)
	// Clean name (replace dots with underscores for tmux)
	sessionName = strings.ReplaceAll(sessionName, ".", "_")
	return sessionName
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

		// Validate that it has command or panels, not both
		if window.Command != "" && len(window.Panels) > 0 {
			log.Printf("Warning: window '%s' has both 'command' and 'panels', using 'panels'", window.Name)
		}

		// Validate layout if there are multiple panels
		if len(window.Panels) > 1 && window.Layout != "horizontal" && window.Layout != "vertical" {
			return fmt.Errorf("window '%s' has multiple panels but invalid layout: '%s' (use 'horizontal' or 'vertical')", window.Name, window.Layout)
		}
	}

	return nil
}
