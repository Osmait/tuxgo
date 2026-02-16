package tmux

import (
	"os"
	"testing"

	"github.com/josesaulburgos/tuxgo/internal/config"
)

func TestGetSessionName(t *testing.T) {
	tests := []struct {
		workDir string
		want    string
	}{
		{"/home/user/my-project", "my-project"},
		{"/home/user/my.dotted.project", "my_dotted_project"},
		{"/home/user/simple", "simple"},
		{"/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.workDir, func(t *testing.T) {
			got := GetSessionName(tt.workDir)
			if got != tt.want {
				t.Errorf("GetSessionName(%q) = %q, want %q", tt.workDir, got, tt.want)
			}
		})
	}
}

func TestIsInsideTmux(t *testing.T) {
	// Save and clear TMUX env
	orig := os.Getenv("TMUX")
	defer os.Setenv("TMUX", orig)

	os.Setenv("TMUX", "")
	if IsInsideTmux() {
		t.Error("expected false when TMUX is empty")
	}

	os.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	if !IsInsideTmux() {
		t.Error("expected true when TMUX is set")
	}
}

func TestValidateConfig_Nil(t *testing.T) {
	err := ValidateConfig(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestValidateConfig_NoWindows(t *testing.T) {
	cfg := &config.ProjectConfig{}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for no windows")
	}
}

func TestValidateConfig_WindowWithoutName(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{Command: "echo hello"},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for window without name")
	}
}

func TestValidateConfig_WindowWithoutConfig(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{Name: "empty"},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for window without command/panels/root")
	}
}

func TestValidateConfig_ValidSimpleWindow(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{Name: "editor", Command: "nvim ."},
		},
	}
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateConfig_InvalidPanelLayout(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{
				Name:   "dev",
				Layout: "invalid",
				Panels: []string{"cmd1", "cmd2"},
			},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid layout")
	}
}

func TestValidateConfig_TooManyChildren(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{
				Name: "workspace",
				Root: &config.PanelConfig{
					Split: "horizontal",
					Children: []config.PanelConfig{
						{Command: "a"},
						{Command: "b"},
						{Command: "c"},
					},
				},
			},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for more than 2 children")
	}
}

func TestValidateConfig_InvalidSplitDirection(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{
				Name: "workspace",
				Root: &config.PanelConfig{
					Split: "diagonal",
					Children: []config.PanelConfig{
						{Command: "a"},
						{Command: "b"},
					},
				},
			},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid split direction")
	}
}

func TestValidateConfig_ValidHierarchical(t *testing.T) {
	cfg := &config.ProjectConfig{
		Windows: []config.WindowConfig{
			{
				Name: "workspace",
				Root: &config.PanelConfig{
					Split: "horizontal",
					Children: []config.PanelConfig{
						{Command: "nvim ."},
						{
							Split: "vertical",
							Children: []config.PanelConfig{
								{Command: "htop"},
								{Command: "watch date"},
							},
						},
					},
				},
			},
		},
	}
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
