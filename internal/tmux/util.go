package tmux

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsInsideTmux returns true if the current process is running inside a tmux session
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// GetSessionName generates a session name based on the working directory.
// Replaces dots with underscores for tmux compatibility.
func GetSessionName(workDir string) string {
	sessionName := filepath.Base(workDir)
	sessionName = strings.ReplaceAll(sessionName, ".", "_")
	return sessionName
}

// ListSessions returns a list of all active tmux session names
func ListSessions() ([]string, error) {
	args := append(baseArgs(), "list-sessions", "-F", "#{session_name}")
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
	session := &Session{Name: sessionName}
	return session.Attach()
}
