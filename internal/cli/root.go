package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Osmait/tuxgo/internal/config"
	"github.com/Osmait/tuxgo/internal/history"
	"github.com/Osmait/tuxgo/internal/matcher"
	"github.com/Osmait/tuxgo/internal/tmux"
	"github.com/Osmait/tuxgo/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "tuxgo [directory]",
	Short: "A tmux session manager",
	Long: `TuxGo automatically creates and configures tmux sessions
based on YAML configuration files. Define your window layouts,
panel splits, and startup commands once, and TuxGo sets everything
up for you.

If a directory name is provided, TuxGo will search for it in your
history of previously used directories using fuzzy matching.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	h, err := history.Load()
	if err != nil {
		log.Fatalf("Error loading history: %v", err)
	}
	defer h.Close()

	var currentDir string

	if len(args) > 0 {
		dirArg := args[0]
		resolvedDir, err := resolveDirectory(dirArg, h)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		currentDir = resolvedDir
	} else {
		currentDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current directory: %v", err)
		}
	}

	cfg, err := config.Load(currentDir)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	var projectConfig *config.ProjectConfig

	if cfg != nil {
		projectConfig = matcher.FindMatchingProject(cfg, currentDir)
	}

	if projectConfig == nil && cfg != nil {
		projectConfig = matcher.GetDefaultConfig(cfg)
	}

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

	if err := tmux.ValidateConfig(projectConfig); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	sessionName := tmux.GetSessionName(currentDir)
	session := &tmux.Session{
		Name:    sessionName,
		WorkDir: currentDir,
	}

	if session.HasSession() {
		fmt.Printf("Connecting to existing tmux session '%s'...\n", sessionName)
	} else {
		fmt.Printf("Creating new tmux session '%s'...\n", sessionName)

		firstWindow := projectConfig.Windows[0]
		firstCommand := firstWindow.Command
		if len(firstWindow.Panels) > 0 || firstWindow.Root != nil {
			firstCommand = ""
		}

		if err := session.CreateSession(firstWindow.Name, firstCommand); err != nil {
			log.Fatalf("Error creating session: %v", err)
		}

		if len(firstWindow.Panels) > 0 || firstWindow.Root != nil {
			if err := session.SetupPanels(firstWindow); err != nil {
				log.Fatalf("Error creating panels for first window '%s': %v", firstWindow.Name, err)
			}
		}

		for _, window := range projectConfig.Windows[1:] {
			if err := createWindow(session, window); err != nil {
				log.Fatalf("Error creating window '%s': %v", window.Name, err)
			}
		}

		if err := session.SelectWindow(firstWindow.Name); err != nil {
			log.Printf("Error selecting initial window: %v", err)
		}
	}

	if err := h.Add(currentDir); err != nil {
		log.Printf("Warning: could not update history: %v", err)
	}

	if err := session.Attach(); err != nil {
		log.Fatalf("Error attaching to tmux session: %v", err)
	}
}

func resolveDirectory(pattern string, h *history.History) (string, error) {
	if _, err := os.Stat(pattern); err == nil {
		absPath, err := filepath.Abs(pattern)
		if err != nil {
			return "", fmt.Errorf("error resolving path: %v", err)
		}
		return absPath, nil
	}

	matches, err := h.Search(pattern)
	if err != nil {
		return "", fmt.Errorf("error searching history: %v", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no directory found matching '%s' in history", pattern)
	}

	if len(matches) == 1 {
		return matches[0].Path, nil
	}

	items := make([]tui.DirItem, len(matches))
	for i, m := range matches {
		items[i] = tui.DirItem{
			Path:     m.Path,
			Name:     m.Name,
			UseCount: m.UseCount,
			LastUsed: m.LastUsed,
		}
	}

	selected, ok, err := tui.SelectDirectory(items)
	if err != nil {
		return "", fmt.Errorf("error selecting directory: %v", err)
	}
	if !ok {
		return "", fmt.Errorf("no directory selected")
	}

	return selected, nil
}

func createWindow(session *tmux.Session, window config.WindowConfig) error {
	if len(window.Panels) > 0 || window.Root != nil {
		return session.CreateWindowWithPanels(window)
	}

	if err := session.CreateWindow(window.Name); err != nil {
		return err
	}

	if window.Command != "" {
		return session.SendKeys(window.Name, window.Command)
	}

	return nil
}
