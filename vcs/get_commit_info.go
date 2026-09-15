package vcs

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/git"
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

// extractApiRepository derives the repository from the origin remote's first url.
// It accepts every form git does (scp-style with or without a user, ssh://, https://, http://,
// git://) and resolves ssh host aliases, so `github-nullstone:acme/repo.git` yields the same
// repository as `git@github.com:acme/repo.git`. Anything it cannot turn into a
// host/owner/name is returned as the zero Repo: the caller decides whether a missing
// repository matters, and a half-built one with a bogus url is worse than none.
func extractApiRepository(cfg *config.RemoteConfig) types.Repo {
	if cfg == nil || len(cfg.URLs) == 0 {
		return types.Repo{}
	}
	return parseRemoteRepository(cfg.URLs[0], resolveSshHost)
}

// parseRemoteRepository is the pure part of extractApiRepository. resolveHost is consulted for
// ssh remotes whose host does not look like a real hostname (no dot) so tests can stub it.
func parseRemoteRepository(raw string, resolveHost func(alias string) string) types.Repo {
	u, err := git.ParseRemote(strings.TrimSpace(raw))
	if err != nil || u == nil {
		return types.Repo{}
	}

	host := u.Hostname()
	if isSshRemote(u) && !strings.Contains(host, ".") && resolveHost != nil {
		if resolved := resolveHost(host); resolved != "" {
			host = resolved
		}
	}

	path := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	segments := strings.Split(path, "/")
	if host == "" || len(segments) < 2 {
		return types.Repo{}
	}
	owner := strings.Join(segments[:len(segments)-1], "/")
	name := segments[len(segments)-1]
	if owner == "" || name == "" {
		return types.Repo{}
	}

	repo := types.Repo{
		Host:  host,
		Owner: owner,
		Name:  name,
		Url:   fmt.Sprintf("https://%s/%s/%s", host, owner, name),
	}
	repo.InferVcsProvider()
	return repo
}

// isSshRemote reports whether the parsed remote goes over ssh: git.ParseRemote assigns the
// "git" scheme to scp-style remotes, and ssh:// urls keep their own.
func isSshRemote(u *url.URL) bool {
	return u.Scheme == "git" || u.Scheme == "ssh" || u.Scheme == "git+ssh" || u.Scheme == "ssh+git"
}

// resolveSshHost asks ssh for the real hostname behind a Host alias from ~/.ssh/config.
// It returns "" when ssh is unavailable or the alias resolves to itself.
func resolveSshHost(alias string) string {
	if _, err := exec.LookPath("ssh"); err != nil {
		return ""
	}
	out, err := exec.Command("ssh", "-G", alias).Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "hostname" && fields[1] != alias {
			return fields[1]
		}
	}
	return ""
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
