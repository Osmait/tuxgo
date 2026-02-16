package tmux

import (
	"fmt"
	"log"

	"github.com/Osmait/tuxgo/internal/config"
)

// ValidateConfig validates that the configuration has valid windows
func ValidateConfig(cfg *config.ProjectConfig) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	if len(cfg.Windows) == 0 {
		return fmt.Errorf("configuration has no windows defined")
	}

	for _, window := range cfg.Windows {
		if window.Name == "" {
			return fmt.Errorf("window without name defined")
		}

		// Count configuration types
		hasCommand := window.Command != ""
		hasPanels := len(window.Panels) > 0
		hasRoot := window.Root != nil

		configCount := 0
		if hasCommand {
			configCount++
		}
		if hasPanels {
			configCount++
		}
		if hasRoot {
			configCount++
		}

		if configCount == 0 {
			return fmt.Errorf("window '%s' has no configuration (need 'command', 'panels', or 'root')", window.Name)
		}

		if configCount > 1 {
			log.Printf("Warning: window '%s' has multiple configurations, using 'root' > 'panels' > 'command'", window.Name)
		}

		// Validate legacy layout if using flat panels
		if hasPanels && !hasRoot {
			if len(window.Panels) > 1 && window.Layout != "horizontal" && window.Layout != "vertical" {
				return fmt.Errorf("window '%s' has multiple panels but invalid layout: '%s' (use 'horizontal' or 'vertical')", window.Name, window.Layout)
			}
		}

		// Validate hierarchical root structure
		if hasRoot {
			if err := validatePanelConfig(window.Root, window.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// validatePanelConfig recursively validates a panel configuration
func validatePanelConfig(panel *config.PanelConfig, windowName string) error {
	if panel == nil {
		return nil
	}

	// Validate split direction if children exist
	if len(panel.Children) > 0 {
		if panel.Split != "" && panel.Split != "horizontal" && panel.Split != "vertical" {
			return fmt.Errorf("invalid split direction '%s' in window '%s' (use 'horizontal' or 'vertical')", panel.Split, windowName)
		}

		if len(panel.Children) > 2 {
			return fmt.Errorf("panel in window '%s' has more than 2 children (maximum 2 for binary splits)", windowName)
		}

		// Warn if a container node also has a command (it will be ignored)
		if panel.Command != "" {
			log.Printf("Warning: panel in window '%s' has both 'command' and 'children'; command will be ignored (move it to a child)", windowName)
		}

		// Recursively validate children
		for i := range panel.Children {
			if err := validatePanelConfig(&panel.Children[i], windowName); err != nil {
				return err
			}
		}
	}

	return nil
}
