package vcs

import (
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/url"
	"os/exec"
	"strings"
)

func GetCommitInfo() (types.CommitInfo, error) {
	ci, err := getCommitInfoFromGoGit()
	if err == nil {
		return ci, nil
	}
	// go-git can choke on partial/shallow clones, missing packfiles, and object
	// alternates that the `git` CLI handles natively. Fall back to it before giving up.
	ci2, fallbackErr := getCommitInfoFromGitCLI()
	if fallbackErr != nil {
		// Return the original go-git error; it's usually more descriptive than the
		// CLI's "fatal: ..." text and matches what users have historically seen.
		return types.CommitInfo{}, err
	}
	return ci2, nil
}

func getCommitInfoFromGoGit() (types.CommitInfo, error) {
	ci := types.CommitInfo{}

	repo, err := GetGitRepo()
	if err != nil {
		return ci, err
	} else if repo == nil {
		return ci, nil
	}

	ref, err := repo.Head()
	if err != nil {
		return ci, err
	} else if ref == nil {
		return ci, nil
	}
	ci.BranchName = ref.Name().Short()
	ci.CommitSha = ref.Hash().String()

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return ci, err
	} else if commit == nil {
		return ci, nil
	}
	ci.AuthorEmail = commit.Author.Email
	ci.AuthorUsername = commit.Author.Name
	ci.CommitMessage = commit.Message

	remotes, err := repo.Remotes()
	if err != nil {
		return ci, err
	}
	for _, remote := range remotes {
		rcfg := remote.Config()
		if rcfg.Name == "origin" {
			ci.Repository = extractApiRepository(rcfg)
			break
		}
	}
	ci.InferCommitUrl()

	return ci, nil
}

func extractApiRepository(cfg *config.RemoteConfig) types.Repo {
	repo := types.Repo{}
	if len(cfg.URLs) == 0 {
		return repo
	}

	if strings.HasPrefix(cfg.URLs[0], "git@") {
		// SSH format: git@github.com:org/repo.git
		rest := strings.TrimSuffix(strings.TrimPrefix(cfg.URLs[0], "git@"), ".git")
		parts := strings.SplitN(rest, ":", 2)
		repo.Host = parts[0]
		repoName := strings.SplitN(parts[1], "/", 2)
		repo.Owner = repoName[0]
		repo.Name = repoName[1]
	} else if strings.HasPrefix(cfg.URLs[0], "https://") {
		// HTTPS format: https://github.com/org/repo.git
		u, err := url.Parse(strings.TrimSuffix(cfg.URLs[0], ".git"))
		if err != nil {
			return repo
		}
		repo.Host = u.Host
		repoName := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
		repo.Owner = repoName[0]
		repo.Name = repoName[1]
	}
	repo.Url = fmt.Sprintf("https://%s/%s/%s", repo.Host, repo.Owner, repo.Name)
	repo.InferVcsProvider()

	return repo
}

func getCommitInfoFromGitCLI() (types.CommitInfo, error) {
	ci := types.CommitInfo{}

	if _, err := exec.LookPath("git"); err != nil {
		return ci, fmt.Errorf("git executable not found in PATH: %w", err)
	}

	out, err := runGit("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return ci, err
	}
	if strings.TrimSpace(out) != "true" {
		return ci, nil
	}

	sha, err := runGit("rev-parse", "HEAD")
	if err != nil {
		return ci, err
	}
	ci.CommitSha = strings.TrimSpace(sha)

	if branch, err := runGit("rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		b := strings.TrimSpace(branch)
		if b != "HEAD" { // detached HEAD reports literal "HEAD"; leave BranchName empty
			ci.BranchName = b
		}
	}

	// %an / %ae / %B separated by newlines. %B (raw body) may span multiple lines,
	// so SplitN with n=3 keeps the full message intact in the final part.
	if out, err := runGit("log", "-1", "--format=%an%n%ae%n%B"); err == nil {
		parts := strings.SplitN(out, "\n", 3)
		if len(parts) >= 1 {
			ci.AuthorUsername = parts[0]
		}
		if len(parts) >= 2 {
			ci.AuthorEmail = parts[1]
		}
		if len(parts) >= 3 {
			ci.CommitMessage = parts[2]
		}
	}

	if remoteURL, err := runGit("remote", "get-url", "origin"); err == nil {
		rcfg := &config.RemoteConfig{
			Name: "origin",
			URLs: []string{strings.TrimSpace(remoteURL)},
		}
		ci.Repository = extractApiRepository(rcfg)
	}

	ci.InferCommitUrl()
	return ci, nil
}

func runGit(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s: %s", strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
