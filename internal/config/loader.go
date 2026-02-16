package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigPath returns the path to the global configuration file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "tuxgo", "config.yaml")
}

// GetLocalConfigPath returns the path to the local project configuration file
// Returns empty string if no local config is found
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

// Load loads the YAML configuration file.
// Checks for local config first (.tuxgo.yaml or .tuxgo.yml in workDir),
// falls back to global config if no local config found.
// Returns nil without error if neither file exists.
func Load(workDir string) (*Config, error) {
	// First, try to load local config
	localConfigPath := GetLocalConfigPath(workDir)
	if localConfigPath != "" {
		data, err := os.ReadFile(localConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error reading local configuration file: %v", err)
		}

		config, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("error parsing local YAML: %v", err)
		}

		return config, nil
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

	config, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	return config, nil
}
