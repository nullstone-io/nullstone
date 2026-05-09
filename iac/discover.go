package iac

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"gopkg.in/nullstone-io/nullstone.v0/git"
)

var (
	blankVcsUrl = url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "local/repo",
	}
)

func Discover(dir string, stackName string, w io.Writer) (*iac.ConfigFiles, error) {
	pmr, err := parseIacFiles(dir, stackName)
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

const (
	defaultRepoConfigDirectory = ".nullstone"
)

// hasYamlFile returns true when dir exists and contains at least one .yml/.yaml file
// (subdirectories don't count). An empty directory or one with only non-YAML siblings
// behaves the same as a missing directory — the caller falls back to flat .nullstone/.
func hasYamlFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch filepath.Ext(entry.Name()) {
		case ".yml", ".yaml":
			return true
		}
	}
	return false
}

func parseIacFiles(dir string, stackName string) (*iac.ConfigFiles, error) {
	rootDir, repo, err := git.GetRootDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error looking for repository root directory: %w", err)
	} else if rootDir == "" {
		rootDir = dir
	}

	repoUrl, err := git.GetVcsUrl(repo)
	if err != nil {
		return nil, fmt.Errorf("error trying to discover the repo url of the local repository: %w", err)
	}
	if repoUrl == nil {
		repoUrl = &blankVcsUrl
	}

	repoName := strings.TrimPrefix(repoUrl.Path, "/")

	// Determine config directory: prefer stack-scoped IaC if any YAML lives there, fall back
	// to flat .nullstone/. This must mirror enigma's webhook fetcher (hasStackConfigDir in
	// enigma/engine/activities/fetch_iac_config_files.go) so that detached-mode payloads
	// match what a webhook-triggered sync would have produced.
	configDir := filepath.Join(rootDir, defaultRepoConfigDirectory)
	if stackName != "" {
		stackDir := filepath.Join(rootDir, defaultRepoConfigDirectory, "stacks", stackName)
		if hasYamlFile(stackDir) {
			configDir = stackDir
		}
	}

	pmr, err := iac.ParseConfigDir(repoUrl.String(), repoName, configDir)
	if err != nil {
		return nil, fmt.Errorf("error parsing nullstone IaC files: %w", err)
	}
	return pmr, nil
}
