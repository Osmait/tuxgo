package tmux

import (
	"fmt"
	"os"
	"os/exec"
)

// Session handles tmux session creation and management
type Session struct {
	Name    string
	WorkDir string
}

// baseArgs returns base arguments for all tmux commands (includes -u for UTF-8)
func baseArgs() []string {
	return []string{"-u"}
}

// HasSession checks if a tmux session already exists
func (s *Session) HasSession() bool {
	args := append(baseArgs(), "has-session", "-t", s.Name)
	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	return cmd.Run() == nil
}

// CreateSession creates a new tmux session with the first window
func (s *Session) CreateSession(firstWindowName, firstCommand string) error {
	args := append(baseArgs(), "new-session", "-d", "-s", s.Name, "-n", firstWindowName)

	cmd := exec.Command("tmux", args...)
	cmd.Dir = s.WorkDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating tmux session: %v", err)
	}

	// If there's a command, execute it in the window
	if firstCommand != "" {
		if err := s.SendKeys(firstWindowName, firstCommand); err != nil {
			return fmt.Errorf("error executing command in first window: %v", err)
		}
	}

	return nil
}

// Attach attaches to the tmux session.
// If already inside tmux, uses switch-client instead of attach-session.
func (s *Session) Attach() error {
	var subcmd string
	if IsInsideTmux() {
		subcmd = "switch-client"
	} else {
		subcmd = "attach-session"
	}

	args := append(baseArgs(), subcmd, "-t", s.Name)
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = s.WorkDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error attaching to tmux session: %v", err)
	}

	return nil
}
