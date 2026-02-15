package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
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
