package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveDefaultConfig creates a global configuration file with default values
func SaveDefaultConfig() error {
	configPath := GetConfigPath()

	// Check if it already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists, don't overwrite
	}

	// Create directory if it doesn't exist
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(defaultConfigTemplate), 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %v", err)
	}

	return nil
}

// InitLocalConfig creates a .tuxgo.yaml file in the specified directory
func InitLocalConfig(workDir string) error {
	configPath := filepath.Join(workDir, ".tuxgo.yaml")

	// Check if it already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists: %s", configPath)
	}

	if err := os.WriteFile(configPath, []byte(localConfigTemplate), 0644); err != nil {
		return fmt.Errorf("error writing local configuration file: %v", err)
	}

	return nil
}

var defaultConfigTemplate = `# TuxGo Configuration
# This file defines custom configurations for tmux sessions
#
# Priority (highest to lowest):
#   1. Local config: .tuxgo.yaml or .tuxgo.yml in project directory
#   2. Global config: ~/.config/tuxgo/config.yaml (this file)
#
# Note: Configuration is required. Run 'tuxgo init' to create a local config.

# Default configuration (optional)
# Used when no project matches
# default:
#   windows:
#     - name: editor
#       command: "nvim ."
#     - name: opencode
#       command: "opencode"

# Specific projects (only used in global config)
# Each project has a pattern (glob) compared with the current path
# First match wins
# projects:
#   - pattern: "*/my-go-project"
#     windows:
#       - name: editor
#         command: "nvim ."
#       - name: server
#         layout: horizontal  # or "vertical"
#         panels:
#           - "go run ."
#           - "tail -f logs/app.log"
#
#   - pattern: "*/frontend-*"
#     windows:
#       - name: editor
#         command: "nvim ."
#       - name: dev
#         layout: vertical
#         panels:
#           - "npm run dev"
#           - "npm test"
`

var localConfigTemplate = `# TuxGo Local Configuration
# This file defines tmux windows for this specific project
# Place this file as .tuxgo.yaml in your project root

windows:
  - name: editor
    command: "nvim ."

  # Example: Simple multi-panel layout (flat)
  # - name: dev
  #   layout: horizontal  # or "vertical"
  #   panels:
  #     - "go run ."
  #     - "tail -f logs/app.log"

  # Example: Mixed layout (hierarchical)
  # Creates: ┌───────────────┬──────────┐
  #          │               │   htop   │
  #          │   opencode    ├──────────┤
  #          │               │  ls -la  │
  #          └───────────────┴──────────┘
  # - name: mixed
  #   root:
  #     split: "horizontal"
  #     children:
  #       - command: "opencode"
  #       - split: "vertical"
  #         children:
  #           - command: "htop"
  #           - command: "ls -la"
`
