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

func (f *Finder) Search(pattern string) ([]history.ScoredEntry, error) {
	return f.history.Search(pattern)
}

func (f *Finder) Add(path string) error {
	return f.history.Add(path)
}

func (f *Finder) Close() error {
	return f.history.Close()
}
