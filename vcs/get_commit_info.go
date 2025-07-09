package vcs

import (
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/url"
	"strings"
)

func GetCommitInfo() (types.CommitInfo, error) {
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
