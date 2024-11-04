package iac

import (
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"gopkg.in/nullstone-io/nullstone.v0/git"
	"io"
	"path/filepath"
	"strings"
)

func Discover(dir string, w io.Writer) (*iac.ParseMapResult, error) {
	pmr, err := parseIacFiles(dir)
	if err != nil {
		return nil, err
	}

	// Emit information about detected IaC files
	numFiles := len(pmr.Overrides)
	if pmr.Config != nil {
		numFiles++
	}
	colorstring.Fprintf(w, "[bold]Found %d IaC files[reset]\n", numFiles)
	if cur := pmr.Config; cur != nil {
		relFilename, _ := filepath.Rel(dir, cur.IacContext.Filename)
		fmt.Fprintf(w, "    📂 %s\n", relFilename)
	}
	for _, cur := range pmr.Overrides {
		relFilename, _ := filepath.Rel(dir, cur.IacContext.Filename)
		fmt.Fprintf(w, "    📂 %s\n", relFilename)
	}
	fmt.Fprintln(w)
	return pmr, nil
}

func parseIacFiles(dir string) (*iac.ParseMapResult, error) {
	rootDir, repo, err := git.GetRootDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error looking for repository root directory: %w", err)
	} else if rootDir == "" {
		rootDir = dir
	}

	repoUrl := git.GetVcsUrl(repo)
	repoName := strings.TrimPrefix(repoUrl.Path, "/")
	pmr, err := iac.ParseConfigDir(repoUrl.String(), repoName, filepath.Join(rootDir, ".nullstone"))
	if err != nil {
		return nil, fmt.Errorf("error parsing nullstone IaC files: %w", err)
	}
	return pmr, nil
}
