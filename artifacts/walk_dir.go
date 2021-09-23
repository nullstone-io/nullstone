package artifacts

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func WalkDir(dir string) ([]string, error) {
	filepaths := make([]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			// we don't care about directories
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}
		filepaths = append(filepaths, rel)
		return nil
	})
	return filepaths, err
}
