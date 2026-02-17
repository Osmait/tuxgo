package finder

import (
	"os"
	"testing"
	"time"

	"github.com/Osmait/tuxgo/internal/history"
)

func TestFinderSearch(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	now := time.Now()
	addEntryWithDate(t, h, "/home/user/my-project", "my-project", now, 5)
	addEntryWithDate(t, h, "/home/user/my-app", "my-app", now.Add(-24*time.Hour), 3)
	addEntryWithDate(t, h, "/home/user/other-dir", "other-dir", now, 1)

	f := New(h)

	results, err := f.Search("myproj")
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

func TestFinderSearchMultiple(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	now := time.Now()
	addEntryWithDate(t, h, "/home/user/my-project", "my-project", now, 5)
	addEntryWithDate(t, h, "/home/user/my-app", "my-app", now.Add(-24*time.Hour), 10)

	f := New(h)

	results, err := f.Search("my")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'my', got %d", len(results))
	}
}

func TestFinderSearchNoMatch(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	for i := 0; i < 5; i++ {
		if err := h.Add("/home/user/my-project"); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	f := New(h)

	results, err := f.Search("xyz")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(results))
	}
}

func TestFinderAdd(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	f := New(h)

	if err := f.Add("/home/user/new-project"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := h.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after Add, got %d", len(entries))
	}

	if entries[0].Name != "new-project" {
		t.Errorf("Expected name 'new-project', got %s", entries[0].Name)
	}

	if entries[0].UseCount != 1 {
		t.Errorf("Expected UseCount 1, got %d", entries[0].UseCount)
	}
}

func TestFinderAddExisting(t *testing.T) {
	h := newTestHistory(t)
	defer h.Close()

	f := New(h)

	if err := f.Add("/home/user/my-project"); err != nil {
		t.Fatalf("First Add failed: %v", err)
	}

	for i := 0; i < 4; i++ {
		if err := f.Add("/home/user/my-project"); err != nil {
			t.Fatalf("Add %d failed: %v", i+2, err)
		}
	}

	entries, err := h.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].UseCount != 5 {
		t.Errorf("Expected UseCount 5, got %d", entries[0].UseCount)
	}
}

func newTestHistory(t *testing.T) *history.History {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "tuxgo-finder-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tmpDir)
	})

	h, err := history.Load()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	return h
}

func addEntryWithDate(t *testing.T, h *history.History, path, name string, lastUsed time.Time, useCount int) {
	t.Helper()
	for i := 0; i < useCount; i++ {
		if err := h.Add(path); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}
}
