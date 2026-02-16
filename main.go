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
	initFlag := flag.Bool("init", false, "Initialize a local .tuxgo.yaml configuration file")
	flag.Parse()

	// Handle init flag - create local config file
	if *initFlag {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current directory: %v", err)
		}

		if err := InitLocalConfig(currentDir); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Created .tuxgo.yaml configuration file in current directory")
		fmt.Println("Edit it to customize your tmux session")
		return
	}

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
	config, err := LoadConfig(currentDir)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Determine which configuration to use
	var projectConfig *ProjectConfig

	// 1. Search for match in configured projects (for global config)
	if config != nil {
		projectConfig = FindMatchingProject(config, currentDir)
	}

	// 2. If no match, use default from YAML
	if projectConfig == nil && config != nil {
		projectConfig = GetDefaultConfig(config)
	}

	// 3. If no configuration found, show error and exit
	if projectConfig == nil {
		fmt.Println("No configuration found for this project.")
		fmt.Println()
		fmt.Println("To get started, create a configuration file:")
		fmt.Println()
		fmt.Println("Option 1 - Local project config (recommended):")
		fmt.Println("  Create .tuxgo.yaml in this directory with your project settings")
		fmt.Println()
		fmt.Println("Option 2 - Global config:")
		fmt.Printf("  Create %s\n", GetConfigPath())
		fmt.Println()
		fmt.Println("Run 'tuxgo --init' to create an example local config.")
		os.Exit(1)
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
	// If it has panels or root layout, create window with layout
	if len(window.Panels) > 0 || window.Root != nil {
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
