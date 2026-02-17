package finder

import (
	"testing"
	"time"

	"github.com/Osmait/tuxgo/internal/history"
)

func TestFinderSearch(t *testing.T) {
	h := &history.History{
		Entries: []history.Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now(), UseCount: 5},
			{Path: "/home/user/my-app", Name: "my-app", LastUsed: time.Now().Add(-24 * time.Hour), UseCount: 3},
			{Path: "/home/user/other-dir", Name: "other-dir", LastUsed: time.Now(), UseCount: 1},
		},
	}

	f := New(h)

	results := f.Search("myproj")

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'myproj', got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "my-project" {
		t.Errorf("Expected 'my-project', got %s", results[0].Name)
	}
}

func TestFinderSearchMultiple(t *testing.T) {
	h := &history.History{
		Entries: []history.Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now(), UseCount: 5},
			{Path: "/home/user/my-app", Name: "my-app", LastUsed: time.Now().Add(-24 * time.Hour), UseCount: 10},
		},
	}

	f := New(h)

	results := f.Search("my")

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'my', got %d", len(results))
	}
}

func TestFinderSearchNoMatch(t *testing.T) {
	h := &history.History{
		Entries: []history.Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now(), UseCount: 5},
		},
	}

	f := New(h)

	results := f.Search("xyz")

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(results))
	}
}

func TestFinderAdd(t *testing.T) {
	h := &history.History{Entries: []history.Entry{}}
	f := New(h)

	f.Add("/home/user/new-project")

	if len(h.Entries) != 1 {
		t.Errorf("Expected 1 entry after Add, got %d", len(h.Entries))
	}

	if h.Entries[0].Name != "new-project" {
		t.Errorf("Expected name 'new-project', got %s", h.Entries[0].Name)
	}

	if h.Entries[0].UseCount != 1 {
		t.Errorf("Expected UseCount 1, got %d", h.Entries[0].UseCount)
	}
}

func TestFinderAddExisting(t *testing.T) {
	h := &history.History{
		Entries: []history.Entry{
			{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now(), UseCount: 5},
		},
	}
	f := New(h)

	f.Add("/home/user/my-project")

	if len(h.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(h.Entries))
	}

	if h.Entries[0].UseCount != 6 {
		t.Errorf("Expected UseCount 6, got %d", h.Entries[0].UseCount)
	}
}
