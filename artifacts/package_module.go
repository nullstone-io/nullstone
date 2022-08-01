package artifacts

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"
)

// PackageModule creates a tar.gz containing the module files
// 'filename' allows a developer to specify where to write the tar.gz
// 'patterns' allows a developer to specify which file patterns are included in the tar.gz
// This is more effective than the built-in tar command because it won't fail if a pattern doesn't match any files
func PackageModule(dir, filename string, patterns []string, excludeFn func(entry GlobEntry) bool) error {
	targzFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating module package: %w", err)
	}
	defer targzFile.Close()
	output := NewTargzer(targzFile, filename)
	defer output.Close()

	entries, err := GlobMany(dir, patterns)
	if err != nil {
		return err
	}

	addEntry := func(entry GlobEntry) error {
		if entry.Path == dir {
			return nil
		}
		if excludeFn != nil && excludeFn(entry) {
			// Skip files that match exclude function
			fmt.Fprintf(os.Stderr, "excluding %q\n", entry.Path)
			return nil
		} else {
			fmt.Fprintf(os.Stderr, "packaging %q\n", entry.Path)
		}
		relPath, err := filepath.Rel(dir, entry.Path)
		if err != nil {
			return fmt.Errorf("error deciphering relative path of tar file: %w", err)
		}
		header, err := tar.FileInfoHeader(entry.Info, relPath)
		if err != nil {
			return fmt.Errorf("error creating file header %s: %w", relPath, err)
		}
		header.Name = relPath
		file, err := os.Open(entry.Path)
		if err != nil {
			return fmt.Errorf("error opening file to package into archive %s: %w", relPath, err)
		}
		defer file.Close()
		return output.AddFile(header, file)
	}

	for _, entry := range entries {
		if err := addEntry(entry); err != nil {
			return err
		}
	}
	return nil
}
