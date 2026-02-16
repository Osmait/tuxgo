package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/josesaulburgos/tuxgo/internal/tmux"
	"github.com/josesaulburgos/tuxgo/internal/tui"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active tmux sessions and attach to one",
	Long:    `Shows an interactive list of all active tmux sessions. Use arrow keys to navigate and enter to attach.`,
	Run:     runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	sessions, err := tmux.ListSessions()
	if err != nil {
		log.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No active tmux sessions found.")
		return
	}

	// Use TUI to select session
	selectedSession, ok, err := tui.SelectSession(sessions)
	if err != nil {
		log.Fatalf("TUI error: %v", err)
	}

	if !ok {
		fmt.Println("No session selected.")
		return
	}

	// Attach to selected session
	if err := tmux.AttachToSession(selectedSession); err != nil {
		log.Fatalf("Error attaching to session: %v", err)
	}
}
