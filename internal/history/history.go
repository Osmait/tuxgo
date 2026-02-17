package history

import (
	"database/sql"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Entry struct {
	ID       int64
	Path     string
	Name     string
	LastUsed time.Time
	UseCount int
}

type ScoredEntry struct {
	Entry
	Score float64
}

type History struct {
	db *sql.DB
}

func Load() (*History, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}

	if err := migrateFromYAML(db); err != nil {
		db.Close()
		return nil, err
	}

	return &History{db: db}, nil
}

func (h *History) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

func (h *History) Add(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	name := filepath.Base(absPath)
	now := time.Now()

	_, err = h.db.Exec(`
		INSERT INTO history (path, name, last_used, use_count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(path) DO UPDATE SET
			last_used = excluded.last_used,
			use_count = use_count + 1
	`, absPath, name, now)

	return err
}

func (h *History) Remove(path string) error {
	absPath, _ := filepath.Abs(path)
	_, err := h.db.Exec(`DELETE FROM history WHERE path = ?`, absPath)
	return err
}

func (h *History) GetAll() ([]Entry, error) {
	rows, err := h.db.Query(`
		SELECT id, path, name, last_used, use_count
		FROM history
		ORDER BY last_used DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Path, &e.Name, &e.LastUsed, &e.UseCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (h *History) Search(pattern string) ([]ScoredEntry, error) {
	entries, err := h.GetAll()
	if err != nil {
		return nil, err
	}

	if pattern == "" {
		return h.scoreAndSort(entries, ""), nil
	}

	lowerPattern := strings.ToLower(pattern)
	var results []Entry

	for _, e := range entries {
		name := strings.ToLower(e.Name)
		if fuzzyMatch(lowerPattern, name) {
			results = append(results, e)
		}
	}

	return h.scoreAndSort(results, lowerPattern), nil
}

func (h *History) scoreAndSort(entries []Entry, pattern string) []ScoredEntry {
	results := make([]ScoredEntry, len(entries))
	for i, e := range entries {
		score := calculateScore(pattern, strings.ToLower(e.Name), e)
		results[i] = ScoredEntry{Entry: e, Score: score}
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
