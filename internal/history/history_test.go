package history

import (
	"os"
	"testing"
	"time"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		pattern string
		target  string
		want    bool
	}{
		{"myproj", "my-project", true},
		{"myproj", "my_project", true},
		{"myproj", "myproject", true},
		{"myproj", "MyProject", true},
		{"mp", "myproject", true},
		{"abc", "xyz", false},
		{"longpattern", "short", false},
		{"", "anything", true},
		{"test", "tset", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.target, func(t *testing.T) {
			got := fuzzyMatch(tt.pattern, tt.target)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.pattern, tt.target, got, tt.want)
			}
		})
	}
}

func TestFuzzyMatchScore(t *testing.T) {
	tests := []struct {
		pattern string
		target  string
		wantMin float64
		wantOK  bool
	}{
		{"test", "test", 15, true},
		{"test", "testing", 15, true},
		{"abc", "xyz", 0, false},
		{"mp", "myproject", 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.target, func(t *testing.T) {
			got := fuzzyMatchScore(tt.pattern, tt.target)
			if tt.wantOK {
				if got < tt.wantMin {
					t.Errorf("fuzzyMatchScore(%q, %q) = %v, want >= %v", tt.pattern, tt.target, got, tt.wantMin)
				}
			} else {
				if got != 0 {
					t.Errorf("fuzzyMatchScore(%q, %q) = %v, want 0", tt.pattern, tt.target, got)
				}
			}
		})
	}
}

func TestHistoryAdd(t *testing.T) {
	h := &History{Entries: []Entry{}}

	h.Add("/home/user/my-project")

	if len(h.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Entries))
	}

	if h.Entries[0].Name != "my-project" {
		t.Errorf("Expected name 'my-project', got %s", h.Entries[0].Name)
	}

	if h.Entries[0].UseCount != 1 {
		t.Errorf("Expected UseCount 1, got %d", h.Entries[0].UseCount)
	}
}

func TestHistoryAddExisting(t *testing.T) {
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/my-project", Name: "my-project", UseCount: 5},
		},
	}

	h.Add("/home/user/my-project")

	if len(h.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Entries))
	}

	if h.Entries[0].UseCount != 6 {
		t.Errorf("Expected UseCount 6, got %d", h.Entries[0].UseCount)
	}
}

func TestHistorySearch(t *testing.T) {
	now := time.Now()
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: now, UseCount: 5},
			{Path: "/home/user/my-app", Name: "my-app", LastUsed: now.Add(-24 * time.Hour), UseCount: 3},
			{Path: "/home/user/other-dir", Name: "other-dir", LastUsed: now, UseCount: 1},
		},
	}

	results := h.Search("myproj")

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'myproj', got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "my-project" {
		t.Errorf("Expected 'my-project', got %s", results[0].Name)
	}
}

func TestHistorySearchMultiple(t *testing.T) {
	now := time.Now()
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: now, UseCount: 5},
			{Path: "/home/user/my-app", Name: "my-app", LastUsed: now.Add(-24 * time.Hour), UseCount: 10},
		},
	}

	results := h.Search("my")

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'my', got %d", len(results))
	}

	if results[0].Score <= results[1].Score {
		t.Errorf("Results should be sorted by score descending")
	}
}

func TestHistorySearchNoMatch(t *testing.T) {
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/my-project", Name: "my-project", UseCount: 5},
		},
	}

	results := h.Search("xyz")

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(results))
	}
}

func TestHistorySearchEmptyPattern(t *testing.T) {
	now := time.Now()
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/project-a", Name: "project-a", LastUsed: now, UseCount: 5},
			{Path: "/home/user/project-b", Name: "project-b", LastUsed: now.Add(-48 * time.Hour), UseCount: 10},
		},
	}

	results := h.Search("")

	if len(results) != 2 {
		t.Errorf("Expected 2 results for empty pattern, got %d", len(results))
	}
}

func TestHistoryRemove(t *testing.T) {
	h := &History{
		Entries: []Entry{
			{Path: "/home/user/my-project", Name: "my-project"},
			{Path: "/home/user/other", Name: "other"},
		},
	}

	h.Remove("/home/user/my-project")

	if len(h.Entries) != 1 {
		t.Errorf("Expected 1 entry after remove, got %d", len(h.Entries))
	}

	if len(h.Entries) > 0 && h.Entries[0].Name != "other" {
		t.Errorf("Expected remaining entry to be 'other', got %s", h.Entries[0].Name)
	}
}

func TestCalculateRecencyScore(t *testing.T) {
	tests := []struct {
		name     string
		lastUsed time.Time
		wantMin  float64
		wantMax  float64
	}{
		{"today", time.Now(), 99, 101},
		{"yesterday", time.Now().Add(-24 * time.Hour), 99, 101},
		{"week_ago", time.Now().Add(-7 * 24 * time.Hour), 10, 20},
		{"month_ago", time.Now().Add(-30 * 24 * time.Hour), 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRecencyScore(tt.lastUsed)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateRecencyScore() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateFrequencyScore(t *testing.T) {
	tests := []struct {
		useCount int
		wantMin  float64
	}{
		{0, -1},
		{1, 0},
		{5, 20},
		{10, 40},
		{100, 80},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calculateFrequencyScore(tt.useCount)
			if tt.useCount == 0 {
				if got != 0 {
					t.Errorf("calculateFrequencyScore(0) = %v, want 0", got)
				}
			} else if got < tt.wantMin {
				t.Errorf("calculateFrequencyScore(%d) = %v, want >= %v", tt.useCount, got, tt.wantMin)
			}
		})
	}
}

func TestHistorySaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tuxgo-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	h1 := &History{
		Entries: []Entry{
			{Path: "/home/user/project", Name: "project", LastUsed: time.Now(), UseCount: 5},
		},
	}

	if err := h1.Save(); err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	h2, err := Load()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(h2.Entries) != 1 {
		t.Errorf("Expected 1 entry after load, got %d", len(h2.Entries))
	}

	if h2.Entries[0].Path != "/home/user/project" {
		t.Errorf("Expected path '/home/user/project', got %s", h2.Entries[0].Path)
	}
}

func TestLoadNoFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tuxgo-history-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	h, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error when file doesn't exist: %v", err)
	}

	if h == nil {
		t.Error("Load() returned nil history")
	}

	if len(h.Entries) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(h.Entries))
	}
}
