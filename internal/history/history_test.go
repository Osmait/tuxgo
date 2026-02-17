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
	h := newTestHistory(t)
	defer h.Close()

	if err := h.Add("/home/user/my-project"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := h.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Name != "my-project" {
		t.Errorf("Expected name 'my-project', got %s", entries[0].Name)
	}

	if entries[0].UseCount != 1 {
		t.Errorf("Expected UseCount 1, got %d", entries[0].UseCount)
	}
}

func TestHistoryAddExisting(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	if err := h.Add("/home/user/my-project"); err != nil {
		t.Fatalf("First Add failed: %v", err)
	}

	if err := h.Add("/home/user/my-project"); err != nil {
		t.Fatalf("Second Add failed: %v", err)
	}

	entries, err := h.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].UseCount != 2 {
		t.Errorf("Expected UseCount 2, got %d", entries[0].UseCount)
	}
}

func TestHistorySearch(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	now := time.Now()
	insertTestEntry(t, h, "/home/user/my-project", "my-project", now, 5)
	insertTestEntry(t, h, "/home/user/my-app", "my-app", now.Add(-24*time.Hour), 3)
	insertTestEntry(t, h, "/home/user/other-dir", "other-dir", now, 1)

	results, err := h.Search("myproj")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'myproj', got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "my-project" {
		t.Errorf("Expected 'my-project', got %s", results[0].Name)
	}
}

func TestHistorySearchMultiple(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	now := time.Now()
	insertTestEntry(t, h, "/home/user/my-project", "my-project", now, 5)
	insertTestEntry(t, h, "/home/user/my-app", "my-app", now.Add(-24*time.Hour), 10)

	results, err := h.Search("my")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'my', got %d", len(results))
	}

	if results[0].Score <= results[1].Score {
		t.Errorf("Results should be sorted by score descending")
	}
}

func TestHistorySearchNoMatch(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	insertTestEntry(t, h, "/home/user/my-project", "my-project", time.Now(), 5)

	results, err := h.Search("xyz")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(results))
	}
}

func TestHistorySearchEmptyPattern(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	now := time.Now()
	insertTestEntry(t, h, "/home/user/project-a", "project-a", now, 5)
	insertTestEntry(t, h, "/home/user/project-b", "project-b", now.Add(-48*time.Hour), 10)

	results, err := h.Search("")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for empty pattern, got %d", len(results))
	}
}

func TestHistoryRemove(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	insertTestEntry(t, h, "/home/user/my-project", "my-project", time.Now(), 5)
	insertTestEntry(t, h, "/home/user/other", "other", time.Now(), 3)

	if err := h.Remove("/home/user/my-project"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	entries, err := h.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after remove, got %d", len(entries))
	}

	if len(entries) > 0 && entries[0].Name != "other" {
		t.Errorf("Expected remaining entry to be 'other', got %s", entries[0].Name)
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

func newTestHistory(t *testing.T) *History {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "tuxgo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tmpDir)
	})

	h, err := Load()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	return h
}

func insertTestEntry(t *testing.T, h *History, path, name string, lastUsed time.Time, useCount int) {
	t.Helper()

	_, err := h.db.Exec(`
		INSERT INTO history (path, name, last_used, use_count)
		VALUES (?, ?, ?, ?)
	`, path, name, lastUsed, useCount)
	if err != nil {
		t.Fatalf("Failed to insert test entry: %v", err)
	}
}
