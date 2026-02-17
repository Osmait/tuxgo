package history

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Path     string    `yaml:"path"`
	Name     string    `yaml:"name"`
	LastUsed time.Time `yaml:"last_used"`
	UseCount int       `yaml:"use_count"`
}

type History struct {
	Entries []Entry `yaml:"entries"`
}

func GetDataPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".local", "share", "tuxgo", "history.yaml")
}

func ensureDataDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dataDir := filepath.Join(homeDir, ".local", "share", "tuxgo")
	return os.MkdirAll(dataDir, 0755)
}

func Load() (*History, error) {
	path := GetDataPath()
	if path == "" {
		return &History{Entries: []Entry{}}, nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &History{Entries: []Entry{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var h History
	if err := yaml.Unmarshal(data, &h); err != nil {
		return nil, err
	}

	if h.Entries == nil {
		h.Entries = []Entry{}
	}

	return &h, nil
}

func (h *History) Save() error {
	if err := ensureDataDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(h)
	if err != nil {
		return err
	}

	return os.WriteFile(GetDataPath(), data, 0644)
}

func (h *History) Add(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	name := filepath.Base(absPath)
	now := time.Now()

	for i, e := range h.Entries {
		if e.Path == absPath {
			h.Entries[i].LastUsed = now
			h.Entries[i].UseCount++
			return
		}
	}

	h.Entries = append(h.Entries, Entry{
		Path:     absPath,
		Name:     name,
		LastUsed: now,
		UseCount: 1,
	})
}

func (h *History) Remove(path string) {
	absPath, _ := filepath.Abs(path)
	for i, e := range h.Entries {
		if e.Path == absPath {
			h.Entries = append(h.Entries[:i], h.Entries[i+1:]...)
			return
		}
	}
}

type ScoredEntry struct {
	Entry
	Score float64
}

func (h *History) Search(pattern string) []ScoredEntry {
	if pattern == "" {
		return h.getAllSorted()
	}

	lowerPattern := strings.ToLower(pattern)
	var results []ScoredEntry

	for _, e := range h.Entries {
		name := strings.ToLower(e.Name)
		if fuzzyMatch(lowerPattern, name) {
			score := calculateScore(lowerPattern, name, e)
			results = append(results, ScoredEntry{Entry: e, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func (h *History) getAllSorted() []ScoredEntry {
	results := make([]ScoredEntry, len(h.Entries))
	for i, e := range h.Entries {
		results[i] = ScoredEntry{Entry: e, Score: calculateScore("", strings.ToLower(e.Name), e)}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func fuzzyMatch(pattern, target string) bool {
	if pattern == "" {
		return true
	}
	if len(pattern) > len(target) {
		return false
	}

	patternIdx := 0
	for _, char := range target {
		if patternIdx < len(pattern) {
			patternChar := rune(pattern[patternIdx])
			if unicode.ToLower(char) == unicode.ToLower(patternChar) {
				patternIdx++
			}
		}
	}

	return patternIdx == len(pattern)
}

func fuzzyMatchScore(pattern, target string) float64 {
	if pattern == "" {
		return 100
	}
	if len(pattern) > len(target) {
		return 0
	}

	score := 0.0
	patternIdx := 0
	consecutive := 0
	lastMatchIdx := -1

	for i, char := range target {
		if patternIdx >= len(pattern) {
			break
		}

		patternChar := rune(pattern[patternIdx])
		targetChar := unicode.ToLower(char)

		if targetChar == unicode.ToLower(patternChar) {
			if i == 0 {
				score += 15
			} else if lastMatchIdx == i-1 {
				consecutive++
				score += float64(5 + consecutive)
			} else {
				score += 1
				consecutive = 0
			}
			lastMatchIdx = i
			patternIdx++
		}
	}

	if patternIdx < len(pattern) {
		return 0
	}

	return score
}

func calculateScore(pattern, target string, e Entry) float64 {
	fuzzyScore := fuzzyMatchScore(pattern, target)

	recencyScore := calculateRecencyScore(e.LastUsed)

	frequencyScore := calculateFrequencyScore(e.UseCount)

	return (fuzzyScore * 10) + (recencyScore * 5) + (frequencyScore * 2)
}

func calculateRecencyScore(lastUsed time.Time) float64 {
	daysSince := time.Since(lastUsed).Hours() / 24

	if daysSince < 1 {
		return 100
	}

	return math.Max(0, 100/daysSince)
}

func calculateFrequencyScore(useCount int) float64 {
	if useCount <= 0 {
		return 0
	}

	return math.Log(float64(useCount)) * 20
}
