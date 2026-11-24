package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/response"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// TEMPORARY: `envs apps set` against a Nullstone server that still keeps a preview environment's apps as a list.
//
// The command sets what a preview environment follows from a repository (its gitops configs) and turns apps on
// or off (its automations). A server released before that model has neither for a preview environment: there,
// the apps of the environment are one list, each app carries the branch or pull request it tracks, and the
// list is replaced as a whole.
//
// Such a server answers 404 for the gitops configs of an environment, and the API client reads a 404 as an
// empty result, not as an error. Without this file the command would report success there and change nothing.
// So the command asks first (envFollowsTriggers) and, on an older server, does what it did before
// (legacyEnvApps).
//
// Remove this file, and its call in setEnvApps, once every Nullstone server serves env gitops configs.

func envGitopsConfigsPath(client *api.Client, stackId, envId int64) string {
	return fmt.Sprintf("/orgs/%s/stacks/%d/envs/%d/gitops_configs", client.Config.OrgName, stackId, envId)
}

func legacyPreviewAppsPath(client *api.Client, stackId, envId int64) string {
	return fmt.Sprintf("/orgs/%s/stacks/%d/envs/%d/preview_apps", client.Config.OrgName, stackId, envId)
}

// envFollowsTriggers reports whether the server keeps what an environment follows per repository.
// The stack and environment are known to exist when this is asked, so a 404 means the server has no such route.
func envFollowsTriggers(ctx context.Context, client *api.Client, stackId, envId int64) (bool, error) {
	res, err := client.Do(ctx, http.MethodGet, envGitopsConfigsPath(client, stackId, envId), nil, nil, nil)
	if err != nil {
		return false, err
	}
	if err := response.Verify(res); err != nil {
		if response.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	res.Body.Close()
	return true, nil
}

// legacyPreviewApp is one entry of an environment's app list, kept as the server sent it.
// Every field is sent back untouched except the ones this command changes: the server replaces the list with
// what it is sent, so a field this CLI does not know about would otherwise be lost.
type legacyPreviewApp map[string]json.RawMessage

func (a legacyPreviewApp) appId() int64 {
	var id int64
	_ = json.Unmarshal(a["appId"], &id)
	return id
}

func (a legacyPreviewApp) repo() string {
	var repo string
	_ = json.Unmarshal(a["repo"], &repo)
	return repo
}

// follow makes the app track a branch or a pull request. The two exclude each other, so the other is cleared.
func (a legacyPreviewApp) follow(branchName *string, pullRequestId *int64) {
	a["branchName"], _ = json.Marshal(branchName)
	a["pullRequestId"], _ = json.Marshal(pullRequestId)
}

// legacyEnvApps is the app list of one preview environment on a server that keeps one
type legacyEnvApps struct {
	Client  *api.Client
	StackId int64
	EnvId   int64
}

func (l legacyEnvApps) List(ctx context.Context) ([]legacyPreviewApp, error) {
	res, err := l.Client.Do(ctx, http.MethodGet, legacyPreviewAppsPath(l.Client, l.StackId, l.EnvId), nil, nil, nil)
	if err != nil {
		return nil, err
	}
	// not ReadJsonVal: a 404 must not read as an empty list, or Replace would empty the environment
	if err := response.Verify(res); err != nil {
		return nil, err
	}
	defer res.Body.Close()
	apps := make([]legacyPreviewApp, 0)
	if err := json.NewDecoder(res.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("error decoding json body: %w", err)
	}
	return apps, nil
}

func (l legacyEnvApps) Replace(ctx context.Context, apps []legacyPreviewApp) error {
	raw, err := json.Marshal(apps)
	if err != nil {
		return err
	}
	res, err := l.Client.Do(ctx, http.MethodPut, legacyPreviewAppsPath(l.Client, l.StackId, l.EnvId), nil, nil, json.RawMessage(raw))
	if err != nil {
		return err
	}
	if err := response.Verify(res); err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// applyLegacy returns the environment's full app list with the update applied, the apps it matched ordered
// by name, and how many entries it matched. An entry that is not matched is carried through as it is: the server
// replaces the whole list, so leaving one out would remove it from the environment.
//
// Here an app is in the environment by being in the list, so --disabled takes the matched apps out of it and
// only an app already in the list can be matched.
func (u envAppsUpdate) applyLegacy(existing []legacyPreviewApp, apps []types.Application) ([]legacyPreviewApp, []types.Application, int) {
	byId := map[int64]types.Application{}
	for _, app := range apps {
		byId[app.Id] = app
	}

	updated := make([]legacyPreviewApp, 0, len(existing))
	matched := make([]types.Application, 0)
	selectedCount := 0
	for _, previewApp := range existing {
		app, known := byId[previewApp.appId()]
		selected := strings.EqualFold(previewApp.repo(), u.Repo) &&
			(u.AppName == "" || (known && strings.EqualFold(app.Name, u.AppName)))
		if !selected {
			updated = append(updated, previewApp)
			continue
		}
		selectedCount++
		if known {
			matched = append(matched, app)
		}
		if u.Disable {
			continue
		}
		previewApp.follow(u.BranchName, u.PullRequestId)
		updated = append(updated, previewApp)
	}
	sort.SliceStable(matched, func(i, j int) bool { return matched[i].Name < matched[j].Name })
	return updated, matched, selectedCount
}

// setLegacyEnvApps runs `envs apps set` against a server that keeps a preview environment's apps as a list.
// It returns the apps it changed.
func setLegacyEnvApps(ctx context.Context, client *api.Client, stack types.Stack, env types.Environment, apps []types.Application, update envAppsUpdate) ([]types.Application, error) {
	if update.Enable {
		return nil, fmt.Errorf("--enabled is not available yet for environment %q; an app joins a preview environment when the environment is created", env.Name)
	}

	list := legacyEnvApps{Client: client, StackId: stack.Id, EnvId: env.Id}
	existing, err := list.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing the apps of environment %q: %w", env.Name, err)
	}
	updated, matched, selectedCount := update.applyLegacy(existing, apps)
	if selectedCount == 0 {
		return nil, fmt.Errorf("no apps in environment %q come from repo %q%s", env.Name, update.Repo, narrowedByApp(update.AppName))
	}
	if err := list.Replace(ctx, updated); err != nil {
		return nil, fmt.Errorf("error updating the apps of environment %q: %w", env.Name, err)
	}
	return matched, nil
}
