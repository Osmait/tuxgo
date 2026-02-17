package finder

import (
	"github.com/Osmait/tuxgo/internal/history"
)

type Finder struct {
	history *history.History
}

func New(h *history.History) *Finder {
	return &Finder{history: h}
}

func (f *Finder) Search(pattern string) []history.ScoredEntry {
	return f.history.Search(pattern)
}

func (f *Finder) Add(path string) {
	f.history.Add(path)
}

func (f *Finder) Save() error {
	return f.history.Save()
}

func SortByScore(pattern string, entries []history.ScoredEntry) []history.ScoredEntry {
	return entries
}
