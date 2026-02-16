package config

import "gopkg.in/yaml.v3"

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
	Name    string       `yaml:"name"`
	Command string       `yaml:"command,omitempty"` // For simple window (single pane)
	Layout  string       `yaml:"layout,omitempty"`  // "horizontal" or "vertical" (legacy, for flat layouts)
	Panels  []string     `yaml:"panels,omitempty"`  // Commands for multiple panels (legacy, flat layouts)
	Root    *PanelConfig `yaml:"root,omitempty"`    // Root panel for mixed/hierarchical layouts
}

// PanelConfig represents a single panel that can have children (for mixed layouts)
type PanelConfig struct {
	Command  string        `yaml:"command,omitempty"`  // Command to run in this panel
	Split    string        `yaml:"split,omitempty"`    // "horizontal" or "vertical" split for children
	Children []PanelConfig `yaml:"children,omitempty"` // Child panels (up to 2 for binary splits)
}

// Parse parses raw YAML bytes into a Config struct
func Parse(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
