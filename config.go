package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration file structure
type Config struct {
	Default  ProjectConfig   `yaml:"default"`
	Projects []ProjectConfig `yaml:"projects"`
}

// ProjectConfig represents a project configuration
// If Pattern is empty, it's considered the default configuration
type ProjectConfig struct {
	Pattern string         `yaml:"pattern,omitempty"`
	Windows []WindowConfig `yaml:"windows"`
}

// WindowConfig represents a window configuration
type WindowConfig struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command,omitempty"` // For simple window (single panel)
	Layout  string   `yaml:"layout,omitempty"`  // "horizontal" or "vertical"
	Panels  []string `yaml:"panels,omitempty"`  // Commands for multiple panels
}

// GetConfigPath returns the path to the global configuration file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "tuxgo", "config.yaml")
}

// GetLocalConfigPath returns the path to the local project configuration file
func GetLocalConfigPath(workDir string) string {
	// Check for .tuxgo.yaml first
	localConfig := filepath.Join(workDir, ".tuxgo.yaml")
	if _, err := os.Stat(localConfig); err == nil {
		return localConfig
	}

	// Fall back to .tuxgo.yml
	localConfig = filepath.Join(workDir, ".tuxgo.yml")
	if _, err := os.Stat(localConfig); err == nil {
		return localConfig
	}

	return ""
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "tuxgo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating configuration directory: %v", err)
	}

	return nil
}

// LoadConfig loads the YAML configuration file
// Checks for local config first (.tuxgo.yaml or .tuxgo.yml in workDir)
// Falls back to global config if no local config found
// Returns nil without error if neither file exists
func LoadConfig(workDir string) (*Config, error) {
	// First, try to load local config
	localConfigPath := GetLocalConfigPath(workDir)
	if localConfigPath != "" {
		data, err := os.ReadFile(localConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error reading local configuration file: %v", err)
		}

		var config Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("error parsing local YAML: %v", err)
		}

		return &config, nil
	}

	// Fall back to global config
	configPath := GetConfigPath()

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	return &config, nil
}

// SaveDefaultConfig creates a configuration file with default values
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

	defaultConfig := `# TuxGo Configuration
# This file defines custom configurations for tmux sessions
#
# Priority (highest to lowest):
#   1. Local config: .tuxgo.yaml or .tuxgo.yml in project directory
#   2. Global config: ~/.config/tuxgo/config.yaml (this file)
#   3. Hardcoded defaults

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
#           - command: "go run ."
#           - command: "tail -f logs/app.log"
#
#   - pattern: "*/frontend-*"
#     windows:
#       - name: editor
#         command: "nvim ."
#       - name: dev
#         layout: vertical
#         panels:
#           - command: "npm run dev"
#           - command: "npm test"

# For local project configs (.tuxgo.yaml), define windows directly:
# windows:
#   - name: editor
#     command: "nvim ."
#   - name: server
#     layout: horizontal
#     panels:
#       - command: "go run ."
#       - command: "tail -f logs/app.log"
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %v", err)
	}

	return nil
}
