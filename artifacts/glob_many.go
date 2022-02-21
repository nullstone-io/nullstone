package artifacts

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

type GlobEntry struct {
	Pattern string
	Path    string
	Info    fs.FileInfo
}

// GlobMany performs a glob on many patterns and returns a set of entries (path and file info header)
// The resulting set of entries is unique by filepath
func GlobMany(dir string, patterns []string) (map[string]GlobEntry, error) {
	entries := map[string]GlobEntry{}
	for _, pattern := range patterns {
		fullPattern := pattern
		if dir != "" {
			fullPattern = path.Join(dir, pattern)
		}

		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("error globbing pattern %q: %w", fullPattern, err)
		}

		for _, match := range matches {
			if _, ok := entries[match]; ok {
				continue
			}
			info, err := os.Lstat(match)
			if err != nil {
				return nil, fmt.Errorf("error finding file information for %q: %w", match, err)
			}
			entries[match] = GlobEntry{
				Pattern: pattern,
				Path:    match,
				Info:    info,
			}
		}
	}
	return entries, nil
}
