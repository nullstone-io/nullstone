package git

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-git/v5"
	"io"
	"os"
)

// FindGitIgnores will search for patterns in `.gitignore`
// All missing and found patterns are returned
func FindGitIgnores(repo *git.Repository, patterns []string) (found []string, missing []string) {
	found = make([]string, 0)
	missing = make([]string, 0)

	wt, err := repo.Worktree()
	if err != nil {
		return
	}

	filename := wt.Filesystem.Join(".gitignore")
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	existing := map[string]struct{}{}
	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		existing[scanner.Text()] = struct{}{}
	}

	for _, pattern := range patterns {
		if _, ok := existing[pattern]; ok {
			found = append(found, pattern)
		} else {
			missing = append(missing, pattern)
		}
	}
	return
}

// AddGitIgnores will add patterns to `.gitignore`
// Any issues with this are silently ignored
func AddGitIgnores(repo *git.Repository, patterns []string) {
	wt, err := repo.Worktree()
	if err != nil {
		return
	}

	filename := wt.Filesystem.Join(".gitignore")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	for _, pattern := range patterns {
		io.WriteString(file, fmt.Sprintln(pattern))
	}
}
