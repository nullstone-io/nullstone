package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// envAppsServer is a Nullstone API for `envs apps set` tests: one stack (primary), one preview env
// (preview-7) and three apps, two of them from acme/shop.
//
// followsTriggers picks the kind of server:
//   - true: the env follows a trigger per repository (gitops configs) and its apps are turned on and off
//   - false: a server released before that. It has no gitops configs route, and the env's apps are a list
//     that is replaced as a whole
type envAppsServer struct {
	followsTriggers bool
	// previewApps is the env's app list on a server that keeps one, as raw JSON
	previewApps string

	// what the command wrote
	putGitopsConfigs []byte
	putAutomations   []byte
	putPreviewApps   []byte
}

const envAppsLegacyList = `[
  {"id": 501, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 11, "contextKey": "7/primary/api",
   "repo": "acme/shop", "repoUrl": "https://github.com/acme/shop", "branchName": null, "pullRequestId": 42, "fromTheFuture": {"kept": true}},
  {"id": 502, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 12, "contextKey": "7/primary/worker",
   "repo": "acme/shop", "repoUrl": "https://github.com/acme/shop", "branchName": null, "pullRequestId": 42},
  {"id": 503, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 13, "contextKey": "7/primary/web",
   "repo": "acme/site", "repoUrl": "https://github.com/acme/site", "branchName": "main", "pullRequestId": null}
]`

func (s *envAppsServer) start(t *testing.T) api.Config {
	t.Helper()
	send := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}
	app := func(id int64, name, repo string) types.Application {
		return types.Application{Block: types.Block{
			IdModel: types.IdModel{Id: id}, OrgName: "acme", StackId: 1, Name: name,
			Repo: repo, RepoUrl: "https://github.com/" + repo,
		}}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/stacks", func(w http.ResponseWriter, r *http.Request) {
		send(w, []types.Stack{{IdModel: types.IdModel{Id: 1}, OrgName: "acme", Name: "primary"}})
	})
	mux.HandleFunc("/orgs/acme/stacks/1/envs", func(w http.ResponseWriter, r *http.Request) {
		send(w, []types.Environment{{IdModel: types.IdModel{Id: 7}, OrgName: "acme", StackId: 1, Name: "preview-7", Type: types.EnvTypePreview}})
	})
	mux.HandleFunc("/orgs/acme/stacks/1/apps", func(w http.ResponseWriter, r *http.Request) {
		send(w, []types.Application{app(12, "worker", "acme/shop"), app(11, "api", "acme/shop"), app(13, "web", "acme/site")})
	})
	if s.followsTriggers {
		mux.HandleFunc("/orgs/acme/stacks/1/envs/7/gitops_configs", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
				s.putGitopsConfigs, _ = io.ReadAll(r.Body)
				send(w, []types.GitopsConnectionConfig{{GitopsConnectionId: 3, EnvId: 7}})
				return
			}
			send(w, []types.GitopsConnectionConfig{})
		})
		mux.HandleFunc("/orgs/acme/stacks/1/envs/7/automations", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
				s.putAutomations, _ = io.ReadAll(r.Body)
			}
			send(w, []types.WorkspaceAutomation{{BlockId: 11, EnvId: 7, IsEnabled: true, AutoDeploy: true}})
		})
	}
	// Both kinds of server answer the read; only a server that keeps the list accepts a write to it.
	mux.HandleFunc("/orgs/acme/stacks/1/envs/7/preview_apps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			if s.followsTriggers {
				http.NotFound(w, r)
				return
			}
			s.putPreviewApps, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(s.putPreviewApps)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(s.previewApps))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return api.Config{BaseAddress: server.URL, OrgName: "acme"}
}

// run runs `envs apps set --stack=primary --env=preview-7` with the update and returns stdout and stderr
func (s *envAppsServer) run(t *testing.T, update envAppsUpdate) (string, string, error) {
	t.Helper()
	require.NoError(t, update.Validate(), "the flags of the test are not a valid invocation")
	cfg := s.start(t)
	var stdout, stderr bytes.Buffer
	err := setEnvApps(context.Background(), cfg, "primary", "preview-7", update, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

// A CI script that points a preview environment at a branch runs
//
//	nullstone envs apps set --stack=<stack> --env=<env> --repo=<owner/name> --branch=<branch>
//
// It must do the same thing on a server that keeps the environment's apps as a list and on one where the
// environment follows a trigger per repository, with the same output, so a script written against one keeps
// working against the other.
func TestSetEnvApps_FollowBranchOfRepo(t *testing.T) {
	followBranch := envAppsUpdate{Repo: "acme/shop", BranchName: strPtr("feature-x")}
	wantStdout := "api\nworker\n"
	wantStderr := "environment \"preview-7\" now follows branch feature-x of \"acme/shop\"\n"

	t.Run("server where the environment follows a trigger per repository", func(t *testing.T) {
		server := &envAppsServer{followsTriggers: true, previewApps: "[]"}

		stdout, stderr, err := server.run(t, followBranch)

		require.NoError(t, err)
		assert.Equal(t, wantStdout, stdout)
		assert.Equal(t, wantStderr, stderr)
		var triggers []api.SetEnvGitopsTriggerInput
		require.NoError(t, json.Unmarshal(server.putGitopsConfigs, &triggers), "the trigger of the repo was not written")
		require.Len(t, triggers, 1)
		assert.Equal(t, "https://github.com/acme/shop", triggers[0].RepoUrl)
		assert.Equal(t, "feature-x", triggers[0].GitBranch)
		assert.Nil(t, triggers[0].PullRequestId)
		assert.Nil(t, server.putPreviewApps, "the app list is not written on this server")
		assert.Nil(t, server.putAutomations, "following a branch does not turn apps on or off")
	})

	t.Run("server that keeps the apps of the environment as a list", func(t *testing.T) {
		server := &envAppsServer{followsTriggers: false, previewApps: envAppsLegacyList}

		stdout, stderr, err := server.run(t, followBranch)

		require.NoError(t, err)
		assert.Equal(t, wantStdout, stdout)
		assert.Equal(t, wantStderr, stderr)
		require.NotNil(t, server.putPreviewApps, "the app list was not written")
		// Every app of the repo follows the branch and no longer the pull request. The app of the other repo
		// and every field this CLI does not know are sent back as they were.
		assert.JSONEq(t, `[
  {"id": 501, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 11, "contextKey": "7/primary/api",
   "repo": "acme/shop", "repoUrl": "https://github.com/acme/shop", "branchName": "feature-x", "pullRequestId": null, "fromTheFuture": {"kept": true}},
  {"id": 502, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 12, "contextKey": "7/primary/worker",
   "repo": "acme/shop", "repoUrl": "https://github.com/acme/shop", "branchName": "feature-x", "pullRequestId": null},
  {"id": 503, "createdBy": "pat", "orgName": "acme", "stackId": 1, "envId": 7, "appId": 13, "contextKey": "7/primary/web",
   "repo": "acme/site", "repoUrl": "https://github.com/acme/site", "branchName": "main", "pullRequestId": null}
]`, string(server.putPreviewApps))
	})

	t.Run("repo is matched without regard to case", func(t *testing.T) {
		for _, followsTriggers := range []bool{true, false} {
			server := &envAppsServer{followsTriggers: followsTriggers, previewApps: envAppsLegacyList}

			stdout, _, err := server.run(t, envAppsUpdate{Repo: "Acme/Shop", BranchName: strPtr("feature-x")})

			require.NoError(t, err)
			assert.Equal(t, wantStdout, stdout)
		}
	})
}

func TestSetEnvApps_ListServer(t *testing.T) {
	t.Run("follow a pull request", func(t *testing.T) {
		server := &envAppsServer{previewApps: envAppsLegacyList}

		_, stderr, err := server.run(t, envAppsUpdate{Repo: "acme/site", PullRequestId: i64Ptr(9)})

		require.NoError(t, err)
		assert.Equal(t, "environment \"preview-7\" now follows pull request 9 of \"acme/site\"\n", stderr)
		var sent []map[string]any
		require.NoError(t, json.Unmarshal(server.putPreviewApps, &sent))
		require.Len(t, sent, 3)
		assert.Nil(t, sent[2]["branchName"])
		assert.EqualValues(t, 9, sent[2]["pullRequestId"])
		assert.EqualValues(t, 42, sent[0]["pullRequestId"], "apps of other repos are untouched")
	})

	t.Run("disabled takes the apps of the repo out of the list", func(t *testing.T) {
		server := &envAppsServer{previewApps: envAppsLegacyList}

		stdout, stderr, err := server.run(t, envAppsUpdate{Repo: "acme/shop", Disable: true})

		require.NoError(t, err)
		assert.Equal(t, "api\nworker\n", stdout)
		assert.Equal(t, "disabled 2 app(s) from \"acme/shop\" in environment \"preview-7\"\n", stderr)
		var sent []map[string]any
		require.NoError(t, json.Unmarshal(server.putPreviewApps, &sent))
		require.Len(t, sent, 1)
		assert.EqualValues(t, 13, sent[0]["appId"])
	})

	t.Run("disabled narrowed to one app", func(t *testing.T) {
		server := &envAppsServer{previewApps: envAppsLegacyList}

		stdout, _, err := server.run(t, envAppsUpdate{Repo: "acme/shop", AppName: "worker", Disable: true})

		require.NoError(t, err)
		assert.Equal(t, "worker\n", stdout)
		var sent []map[string]any
		require.NoError(t, json.Unmarshal(server.putPreviewApps, &sent))
		require.Len(t, sent, 2)
		assert.EqualValues(t, 11, sent[0]["appId"])
		assert.EqualValues(t, 13, sent[1]["appId"])
	})

	t.Run("repo with no app in the environment writes nothing", func(t *testing.T) {
		server := &envAppsServer{previewApps: envAppsLegacyList}

		_, _, err := server.run(t, envAppsUpdate{Repo: "acme/other", BranchName: strPtr("main")})

		require.EqualError(t, err, `no apps in environment "preview-7" come from repo "acme/other"`)
		assert.Nil(t, server.putPreviewApps)
	})

	t.Run("enabled is refused and writes nothing", func(t *testing.T) {
		server := &envAppsServer{previewApps: envAppsLegacyList}

		_, _, err := server.run(t, envAppsUpdate{Repo: "acme/shop", Enable: true})

		require.ErrorContains(t, err, "--enabled is not available yet")
		assert.Nil(t, server.putPreviewApps)
	})
}

func TestSetEnvApps_TriggerServer(t *testing.T) {
	t.Run("disabled turns the apps of the repo off", func(t *testing.T) {
		server := &envAppsServer{followsTriggers: true, previewApps: "[]"}

		stdout, stderr, err := server.run(t, envAppsUpdate{Repo: "acme/shop", Disable: true})

		require.NoError(t, err)
		assert.Equal(t, "api\nworker\n", stdout)
		assert.Equal(t, "disabled 2 app(s) from \"acme/shop\" in environment \"preview-7\"\n", stderr)
		var sent []api.UpdateWorkspaceAutomationInput
		require.NoError(t, json.Unmarshal(server.putAutomations, &sent))
		assert.Equal(t, []api.UpdateWorkspaceAutomationInput{
			{BlockId: 11, IsEnabled: false, AutoDeploy: true},
			{BlockId: 12, IsEnabled: false, AutoDeploy: false},
		}, sent)
		assert.Nil(t, server.putGitopsConfigs)
		assert.Nil(t, server.putPreviewApps)
	})

	t.Run("repo with no app in the stack writes nothing", func(t *testing.T) {
		server := &envAppsServer{followsTriggers: true, previewApps: "[]"}

		_, _, err := server.run(t, envAppsUpdate{Repo: "acme/other", BranchName: strPtr("main")})

		require.EqualError(t, err, `no apps in stack "primary" come from repo "acme/other"`)
		assert.Nil(t, server.putGitopsConfigs)
	})
}

// The API client reads a 404 as an empty result, so a write that the server has no route for must not be
// reported as done.
func TestEnvAppsUpdate_Apply_WriteWithNoRoute(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := api.Client{Config: api.Config{BaseAddress: server.URL, OrgName: "acme"}}
	stack := types.Stack{IdModel: types.IdModel{Id: 1}, OrgName: "acme", Name: "primary"}
	env := types.Environment{IdModel: types.IdModel{Id: 7}, OrgName: "acme", StackId: 1, Name: "preview-7", Type: types.EnvTypePreview}
	update := envAppsUpdate{Repo: "nullstone-io/backend", BranchName: strPtr("main")}

	_, err := update.apply(context.Background(), &client, stack, env, envAppFixtures())

	require.EqualError(t, err, `error updating what environment "preview-7" follows: the server did not accept the change`)
}

func TestEnvFollowsTriggers(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		want    bool
		wantErr bool
	}{
		{name: "the server has the route", status: http.StatusOK, want: true},
		{name: "the server has no such route", status: http.StatusNotFound, want: false},
		{name: "not allowed to read it", status: http.StatusForbidden, wantErr: true},
		{name: "server error", status: http.StatusInternalServerError, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/orgs/acme/stacks/1/envs/7/gitops_configs", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(`[]`))
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)
			client := api.Client{Config: api.Config{BaseAddress: server.URL, OrgName: "acme"}}

			got, err := envFollowsTriggers(context.Background(), &client, 1, 7)

			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
