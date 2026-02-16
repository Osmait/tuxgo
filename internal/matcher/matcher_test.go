package matcher

import (
	"testing"

	"github.com/Osmait/tuxgo/internal/config"
)

func TestMatchPattern_SimpleWildcard(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.go", "/home/user/project", false},
		{"project", "/home/user/project", true},
		{"proj*", "/home/user/project", true},
		{"*oject", "/home/user/project", true},
		{"other", "/home/user/project", false},
		{"", "/home/user/project", false},
		{"project", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			got := MatchPattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchPattern_DoubleStar(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/project", "/home/user/project", true},
		{"**/project", "/home/user/other/project", true},
		{"**/other", "/home/user/project", false},
		{"/home/**", "/home/user/project", true},
		{"/home/**", "/other/user/project", false},
		{"/home/**/project", "/home/user/project", true},
		{"/home/**/project", "/home/deep/nested/project", true},
		{"/home/**/project", "/home/user/other", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			got := MatchPattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchPattern_TrailingSlash(t *testing.T) {
	got := MatchPattern("project", "/home/user/project/")
	if !got {
		t.Error("expected trailing slash to be normalized")
	}
}

func TestFindMatchingProject(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{
				Pattern: "frontend-*",
				Windows: []config.WindowConfig{{Name: "frontend"}},
			},
			{
				Pattern: "backend",
				Windows: []config.WindowConfig{{Name: "backend"}},
			},
		},
	}

	// Should match first project
	result := FindMatchingProject(cfg, "/home/user/frontend-app")
	if result == nil {
		t.Fatal("expected a match for frontend-app")
	}
	if result.Windows[0].Name != "frontend" {
		t.Errorf("expected 'frontend', got '%s'", result.Windows[0].Name)
	}

	// Should match second project
	result = FindMatchingProject(cfg, "/home/user/backend")
	if result == nil {
		t.Fatal("expected a match for backend")
	}
	if result.Windows[0].Name != "backend" {
		t.Errorf("expected 'backend', got '%s'", result.Windows[0].Name)
	}

	// No match
	result = FindMatchingProject(cfg, "/home/user/other")
	if result != nil {
		t.Error("expected no match for 'other'")
	}

	// Nil config
	result = FindMatchingProject(nil, "/home/user/project")
	if result != nil {
		t.Error("expected nil for nil config")
	}
}

func TestGetDefaultConfig(t *testing.T) {
	// With default windows
	cfg := &config.Config{
		Default: config.ProjectConfig{
			Windows: []config.WindowConfig{{Name: "editor"}},
		},
	}
	result := GetDefaultConfig(cfg)
	if result == nil {
		t.Fatal("expected default config")
	}

	// Empty default
	cfg = &config.Config{}
	result = GetDefaultConfig(cfg)
	if result != nil {
		t.Error("expected nil for empty default")
	}

	// Nil config
	result = GetDefaultConfig(nil)
	if result != nil {
		t.Error("expected nil for nil config")
	}
}
