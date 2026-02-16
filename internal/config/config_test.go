package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_ValidConfig(t *testing.T) {
	yaml := `
default:
  windows:
    - name: editor
      command: "nvim ."
    - name: dev
      layout: horizontal
      panels:
        - "go run ."
        - "tail -f app.log"

projects:
  - pattern: "*/my-project"
    windows:
      - name: code
        command: "code ."
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Default.Windows) != 2 {
		t.Errorf("expected 2 default windows, got %d", len(cfg.Default.Windows))
	}

	if cfg.Default.Windows[0].Name != "editor" {
		t.Errorf("expected window name 'editor', got '%s'", cfg.Default.Windows[0].Name)
	}

	if cfg.Default.Windows[0].Command != "nvim ." {
		t.Errorf("expected command 'nvim .', got '%s'", cfg.Default.Windows[0].Command)
	}

	if cfg.Default.Windows[1].Layout != "horizontal" {
		t.Errorf("expected layout 'horizontal', got '%s'", cfg.Default.Windows[1].Layout)
	}

	if len(cfg.Default.Windows[1].Panels) != 2 {
		t.Errorf("expected 2 panels, got %d", len(cfg.Default.Windows[1].Panels))
	}

	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(cfg.Projects))
	}

	if cfg.Projects[0].Pattern != "*/my-project" {
		t.Errorf("expected pattern '*/my-project', got '%s'", cfg.Projects[0].Pattern)
	}
}

func TestParse_HierarchicalLayout(t *testing.T) {
	yaml := `
default:
  windows:
    - name: workspace
      root:
        split: "horizontal"
        children:
          - command: "nvim ."
          - split: "vertical"
            children:
              - command: "htop"
              - command: "watch date"
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	win := cfg.Default.Windows[0]
	if win.Root == nil {
		t.Fatal("expected root to be set")
	}

	if win.Root.Split != "horizontal" {
		t.Errorf("expected split 'horizontal', got '%s'", win.Root.Split)
	}

	if len(win.Root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(win.Root.Children))
	}

	if win.Root.Children[0].Command != "nvim ." {
		t.Errorf("expected first child command 'nvim .', got '%s'", win.Root.Children[0].Command)
	}

	secondChild := win.Root.Children[1]
	if secondChild.Split != "vertical" {
		t.Errorf("expected second child split 'vertical', got '%s'", secondChild.Split)
	}

	if len(secondChild.Children) != 2 {
		t.Fatalf("expected 2 grandchildren, got %d", len(secondChild.Children))
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	yaml := `
default:
  windows:
    - name: [invalid
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_EmptyConfig(t *testing.T) {
	cfg, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Default.Windows) != 0 {
		t.Errorf("expected 0 windows, got %d", len(cfg.Default.Windows))
	}
}

func TestGetLocalConfigPath_YamlExtension(t *testing.T) {
	dir := t.TempDir()

	// No config exists
	path := GetLocalConfigPath(dir)
	if path != "" {
		t.Errorf("expected empty path, got '%s'", path)
	}

	// Create .tuxgo.yaml
	yamlPath := filepath.Join(dir, ".tuxgo.yaml")
	if err := os.WriteFile(yamlPath, []byte("windows: []"), 0644); err != nil {
		t.Fatal(err)
	}

	path = GetLocalConfigPath(dir)
	if path != yamlPath {
		t.Errorf("expected '%s', got '%s'", yamlPath, path)
	}
}

func TestGetLocalConfigPath_YmlExtension(t *testing.T) {
	dir := t.TempDir()

	// Create .tuxgo.yml (not .yaml)
	ymlPath := filepath.Join(dir, ".tuxgo.yml")
	if err := os.WriteFile(ymlPath, []byte("windows: []"), 0644); err != nil {
		t.Fatal(err)
	}

	path := GetLocalConfigPath(dir)
	if path != ymlPath {
		t.Errorf("expected '%s', got '%s'", ymlPath, path)
	}
}

func TestGetLocalConfigPath_YamlTakesPriority(t *testing.T) {
	dir := t.TempDir()

	// Create both .tuxgo.yaml and .tuxgo.yml
	yamlPath := filepath.Join(dir, ".tuxgo.yaml")
	ymlPath := filepath.Join(dir, ".tuxgo.yml")
	os.WriteFile(yamlPath, []byte("windows: []"), 0644)
	os.WriteFile(ymlPath, []byte("windows: []"), 0644)

	path := GetLocalConfigPath(dir)
	if path != yamlPath {
		t.Errorf("expected .yaml to take priority, got '%s'", path)
	}
}

func TestInitLocalConfig(t *testing.T) {
	dir := t.TempDir()

	err := InitLocalConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should exist
	configPath := filepath.Join(dir, ".tuxgo.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected .tuxgo.yaml to be created")
	}

	// Calling again should fail (already exists)
	err = InitLocalConfig(dir)
	if err == nil {
		t.Fatal("expected error when config already exists")
	}
}

func TestLoad_LocalConfig(t *testing.T) {
	dir := t.TempDir()

	yaml := `
windows:
  - name: editor
    command: "nvim ."
`
	configPath := filepath.Join(dir, ".tuxgo.yaml")
	os.WriteFile(configPath, []byte(yaml), 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected config to be loaded")
	}
}

func TestLoad_NoConfig(t *testing.T) {
	dir := t.TempDir()

	// Override home to prevent loading global config
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg != nil {
		t.Error("expected nil config when no files exist")
	}
}
