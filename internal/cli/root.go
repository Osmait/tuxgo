package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/josesaulburgos/tuxgo/internal/config"
	"github.com/josesaulburgos/tuxgo/internal/matcher"
	"github.com/josesaulburgos/tuxgo/internal/tmux"
)

var rootCmd = &cobra.Command{
	Use:   "tuxgo",
	Short: "A tmux session manager",
	Long: `TuxGo automatically creates and configures tmux sessions
based on YAML configuration files. Define your window layouts,
panel splits, and startup commands once, and TuxGo sets everything
up for you.`,
	Run: runRoot,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Load configuration
	cfg, err := config.Load(currentDir)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Determine which configuration to use
	var projectConfig *config.ProjectConfig

	// 1. Search for match in configured projects (for global config)
	if cfg != nil {
		projectConfig = matcher.FindMatchingProject(cfg, currentDir)
	}

	// 2. If no match, use default from YAML
	if projectConfig == nil && cfg != nil {
		projectConfig = matcher.GetDefaultConfig(cfg)
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
		fmt.Printf("  Create %s\n", config.GetConfigPath())
		fmt.Println()
		fmt.Println("Run 'tuxgo init' to create an example local config.")
		os.Exit(1)
	}

	// Validate configuration
	if err := tmux.ValidateConfig(projectConfig); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create tmux session
	sessionName := tmux.GetSessionName(currentDir)
	session := &tmux.Session{
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
		firstCommand := firstWindow.Command
		// If first window has panels/root, don't send command via CreateSession
		// (SetupPanels will handle everything)
		if len(firstWindow.Panels) > 0 || firstWindow.Root != nil {
			firstCommand = ""
		}

		if err := session.CreateSession(firstWindow.Name, firstCommand); err != nil {
			log.Fatalf("Error creating session: %v", err)
		}

		// Handle first window panels/root if it has them
		if len(firstWindow.Panels) > 0 || firstWindow.Root != nil {
			if err := session.SetupPanels(firstWindow); err != nil {
				log.Fatalf("Error creating panels for first window '%s': %v", firstWindow.Name, err)
			}
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
	if err := session.Attach(); err != nil {
		log.Fatalf("Error attaching to tmux session: %v", err)
	}
}

// createWindow creates a window according to its configuration
func createWindow(session *tmux.Session, window config.WindowConfig) error {
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
