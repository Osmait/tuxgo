package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Parse command line flags
	listFlag := flag.Bool("l", false, "List active tmux sessions and attach to one")
	listLongFlag := flag.Bool("list", false, "List active tmux sessions and attach to one")
	flag.Parse()

	// Handle list flag - show sessions and let user choose
	if *listFlag || *listLongFlag {
		if err := listAndSelectSession(); err != nil {
			log.Fatalf("Error: %v", err)
		}
		return
	}

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		log.Println("Using default behavior...")
		config = nil
	}

	// If configuration doesn't exist, create example file
	if config == nil {
		if err := SaveDefaultConfig(); err != nil {
			log.Printf("Error creating default configuration: %v", err)
		} else {
			fmt.Printf("Example configuration created at: %s\n", GetConfigPath())
		}
	}

	// Determine which configuration to use
	var projectConfig *ProjectConfig

	// 1. Search for match in configured projects
	if config != nil {
		projectConfig = FindMatchingProject(config, currentDir)
	}

	// 2. If no match, use default from YAML
	if projectConfig == nil && config != nil {
		projectConfig = GetDefaultConfig(config)
	}

	// 3. If no configuration, use hardcoded default
	if projectConfig == nil {
		projectConfig = getHardcodedDefault()
	}

	// Validate configuration
	if err := ValidateConfig(projectConfig); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create tmux session
	sessionName := GetSessionName(currentDir)
	session := &TmuxSession{
		Name:    sessionName,
		WorkDir: currentDir,
	}

	// Check if session already exists
	if session.HasSession() {
		fmt.Printf("Connecting to existing tmux session '%s'...\n", sessionName)
	} else {
		fmt.Printf("Creating new tmux session '%s'...\n", sessionName)

		// Create session with first window
		firstWindow := projectConfig.Windows[0]
		if err := session.CreateSession(firstWindow.Name, firstWindow.Command); err != nil {
			log.Fatalf("Error creating session: %v", err)
		}

		// Create additional windows
		for _, window := range projectConfig.Windows[1:] {
			if err := createWindow(session, window); err != nil {
				log.Fatalf("Error creating window '%s': %v", window.Name, err)
			}
		}

		// Select first window
		if err := session.SelectWindow(firstWindow.Name); err != nil {
			log.Printf("Error selecting initial window: %v", err)
		}
	}

	// Attach to session
	if err := session.AttachSession(); err != nil {
		log.Fatalf("Error attaching to tmux session: %v", err)
	}
}

// listAndSelectSession lists all active sessions using an interactive TUI
func listAndSelectSession() error {
	sessions, err := ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No active tmux sessions found.")
		return nil
	}

	// Use TUI to select session
	selectedSession, ok, err := SelectSessionTUI(sessions)
	if err != nil {
		return fmt.Errorf("TUI error: %v", err)
	}

	if !ok {
		fmt.Println("No session selected.")
		return nil
	}

	// Attach to selected session
	if err := AttachToSession(selectedSession); err != nil {
		return fmt.Errorf("error attaching to session: %v", err)
	}

	return nil
}

// createWindow creates a window according to its configuration
func createWindow(session *TmuxSession, window WindowConfig) error {
	// If it has panels, create window with layout
	if len(window.Panels) > 0 {
		return session.CreateWindowWithPanels(window)
	}

	// If it only has command, create simple window
	if err := session.CreateWindow(window.Name); err != nil {
		return err
	}

	if window.Command != "" {
		return session.SendKeys(window.Name, window.Command)
	}

	return nil
}

// getHardcodedDefault returns the hardcoded default configuration
// Used when there's no configuration file
func getHardcodedDefault() *ProjectConfig {
	return &ProjectConfig{
		Windows: []WindowConfig{
			{
				Name:    "editor",
				Command: "nvim .",
			},
			{
				Name:    "opencode",
				Command: "opencode",
			},
		},
	}
}
