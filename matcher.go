package main

import (
	"path/filepath"
	"strings"
)

// FindMatchingProject searches for the first project whose pattern matches the current path
// Returns nil if no match is found
func FindMatchingProject(config *Config, currentPath string) *ProjectConfig {
	if config == nil {
		return nil
	}

	for _, project := range config.Projects {
		if MatchPattern(project.Pattern, currentPath) {
			return &project
		}
	}

	return nil
}

// MatchPattern checks if a path matches a glob pattern
// Supports wildcards like *, **, ?
func MatchPattern(pattern, path string) bool {
	if pattern == "" {
		return false
	}

	// Normalize path (remove trailing slash)
	path = strings.TrimSuffix(path, string(filepath.Separator))

	// If pattern contains **, we need special handling
	// since filepath.Match doesn't support **
	if strings.Contains(pattern, "**") {
		return matchDoubleStar(pattern, path)
	}

	// For simple patterns, use filepath.Match
	// We need to compare either the final directory name or the full path
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	// Try matching with the base directory name
	base := filepath.Base(path)
	matched, err = filepath.Match(pattern, base)
	if err == nil && matched {
		return true
	}

	return false
}

// matchDoubleStar handles patterns with ** (matches multiple directory levels)
func matchDoubleStar(pattern, path string) bool {
	// Special case: ** at the beginning
	if strings.HasPrefix(pattern, "**/") {
		suffix := pattern[3:] // Remove "**/"
		return strings.HasSuffix(path, suffix) || strings.Contains(path, "/"+suffix)
	}

	// Special case: ** at the end
	if strings.HasSuffix(pattern, "/**") {
		prefix := pattern[:len(pattern)-3] // Remove "/**"
		return strings.HasPrefix(path, prefix)
	}

	// ** in the middle - split and search
	parts := strings.Split(pattern, "/**/")
	if len(parts) != 2 {
		// Multiple ** not supported in simple implementation
		return false
	}

	prefix := parts[0]
	suffix := parts[1]

	if !strings.HasPrefix(path, prefix) {
		return false
	}

	pathWithoutPrefix := path[len(prefix):]
	return strings.HasSuffix(pathWithoutPrefix, suffix) ||
		strings.Contains(pathWithoutPrefix, "/"+suffix)
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig(config *Config) *ProjectConfig {
	if config == nil || len(config.Default.Windows) == 0 {
		return nil
	}
	return &config.Default
}
