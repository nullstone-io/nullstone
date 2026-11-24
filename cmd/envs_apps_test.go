package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func strPtr(s string) *string { return &s }
func i64Ptr(i int64) *int64   { return &i }

func envAppFixtures() []types.Application {
	app := func(id int64, name, repo string) types.Application {
		repoUrl := ""
		if repo != "" {
			repoUrl = "https://github.com/" + repo
		}
		return types.Application{Block: types.Block{IdModel: types.IdModel{Id: id}, Name: name, Repo: repo, RepoUrl: repoUrl}}
	}
	return []types.Application{
		app(2, "worker", "nullstone-io/backend"),
		app(1, "api", "nullstone-io/backend"),
		app(3, "web", "nullstone-io/frontend"),
		app(4, "cron", ""),
	}
}

func appNames(apps []types.Application) []string {
	names := make([]string, len(apps))
	for i, app := range apps {
		names[i] = app.Name
	}
	return names
}

func TestEnvAppsUpdate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		update  envAppsUpdate
		wantErr string
	}{
		{name: "branch", update: envAppsUpdate{BranchName: strPtr("main")}},
		{name: "pull request", update: envAppsUpdate{PullRequestId: i64Ptr(42)}},
		{name: "enabled", update: envAppsUpdate{Enable: true}},
		{name: "disabled", update: envAppsUpdate{Disable: true}},
		{name: "enabled narrowed to an app", update: envAppsUpdate{Enable: true, AppName: "api"}},
		{name: "disabled narrowed to an app", update: envAppsUpdate{Disable: true, AppName: "api"}},
		{
			name:    "nothing to do",
			update:  envAppsUpdate{},
			wantErr: "specify one of --branch, --pull-request, --enabled, or --disabled",
		},
		{
			name:    "branch and pull request",
			update:  envAppsUpdate{BranchName: strPtr("main"), PullRequestId: i64Ptr(42)},
			wantErr: "--branch and --pull-request cannot be used together",
		},
		{
			name:    "enabled and disabled",
			update:  envAppsUpdate{Enable: true, Disable: true},
			wantErr: "--enabled and --disabled cannot be used together",
		},
		{
			name:    "disabled with a branch",
			update:  envAppsUpdate{Disable: true, BranchName: strPtr("main")},
			wantErr: "--enabled and --disabled cannot be combined with --branch or --pull-request",
		},
		{
			name:    "enabled with a pull request",
			update:  envAppsUpdate{Enable: true, PullRequestId: i64Ptr(42)},
			wantErr: "--enabled and --disabled cannot be combined with --branch or --pull-request",
		},
		{
			// The environment follows one branch or pull request per repo, so one app cannot differ
			name:    "branch narrowed to an app",
			update:  envAppsUpdate{BranchName: strPtr("main"), AppName: "api"},
			wantErr: "--app cannot be combined with --branch or --pull-request",
		},
		{
			name:    "pull request narrowed to an app",
			update:  envAppsUpdate{PullRequestId: i64Ptr(42), AppName: "api"},
			wantErr: "--app cannot be combined with --branch or --pull-request",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.update.Repo = "nullstone-io/backend"
			err := test.update.Validate()
			if test.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestEnvAppsUpdate_Select(t *testing.T) {
	tests := []struct {
		name   string
		update envAppsUpdate
		want   []string
	}{
		{
			name:   "every app from the repo, ordered by name",
			update: envAppsUpdate{Repo: "nullstone-io/backend"},
			want:   []string{"api", "worker"},
		},
		{
			name:   "narrowed to an app",
			update: envAppsUpdate{Repo: "nullstone-io/backend", AppName: "worker"},
			want:   []string{"worker"},
		},
		{
			name:   "repo and app are matched case-insensitively",
			update: envAppsUpdate{Repo: "NullStone-IO/BackEnd", AppName: "WORKER"},
			want:   []string{"worker"},
		},
		{
			name:   "unknown repo matches nothing",
			update: envAppsUpdate{Repo: "nullstone-io/missing"},
			want:   []string{},
		},
		{
			name:   "an app from another repo matches nothing",
			update: envAppsUpdate{Repo: "nullstone-io/backend", AppName: "web"},
			want:   []string{},
		},
		{
			name:   "an empty repo does not select apps that have no repo",
			update: envAppsUpdate{Repo: "nullstone-io/"},
			want:   []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apps := envAppFixtures()

			got := test.update.Select(apps)

			assert.Equal(t, test.want, appNames(got))
			// the input is not reordered
			assert.Equal(t, appNames(envAppFixtures()), appNames(apps))
		})
	}
}

func TestEnvAppsUpdate_Trigger(t *testing.T) {
	backend := envAppsUpdate{Repo: "nullstone-io/backend"}.Select(envAppFixtures())

	t.Run("branch", func(t *testing.T) {
		update := envAppsUpdate{Repo: "nullstone-io/backend", BranchName: strPtr("feat/x")}

		got, err := update.Trigger(backend)

		require.NoError(t, err)
		assert.Equal(t, api.SetEnvGitopsTriggerInput{
			RepoUrl:       "https://github.com/nullstone-io/backend",
			GitopsTrigger: types.GitopsTrigger{GitBranch: "feat/x"},
		}, got)
		assert.Equal(t, "branch feat/x", update.describeTrigger())
	})

	t.Run("pull request", func(t *testing.T) {
		update := envAppsUpdate{Repo: "nullstone-io/backend", PullRequestId: i64Ptr(99)}

		got, err := update.Trigger(backend)

		require.NoError(t, err)
		assert.Equal(t, api.SetEnvGitopsTriggerInput{
			RepoUrl:       "https://github.com/nullstone-io/backend",
			GitopsTrigger: types.GitopsTrigger{PullRequestId: i64Ptr(99)},
		}, got)
		assert.Equal(t, "pull request 99", update.describeTrigger())
	})

	t.Run("apps without a repository url cannot be followed", func(t *testing.T) {
		update := envAppsUpdate{Repo: "nullstone-io/backend", BranchName: strPtr("feat/x")}
		noUrl := []types.Application{{Block: types.Block{Name: "api", Repo: "nullstone-io/backend"}}}

		_, err := update.Trigger(noUrl)

		require.Error(t, err)
	})
}

func TestEnvAppsUpdate_Automations(t *testing.T) {
	backend := envAppsUpdate{Repo: "nullstone-io/backend"}.Select(envAppFixtures())
	current := []types.WorkspaceAutomation{
		{BlockId: 1, IsEnabled: true, AutoDeploy: true},
		{BlockId: 3, IsEnabled: true, AutoDeploy: true},
		// worker (2) has no automation yet
	}

	t.Run("disable keeps auto-deploy and leaves other apps out of the write", func(t *testing.T) {
		got := envAppsUpdate{Repo: "nullstone-io/backend", Disable: true}.Automations(backend, current)

		assert.Equal(t, []api.UpdateWorkspaceAutomationInput{
			{BlockId: 1, IsEnabled: false, AutoDeploy: true},
			{BlockId: 2, IsEnabled: false, AutoDeploy: false},
		}, got)
	})

	t.Run("enable keeps auto-deploy", func(t *testing.T) {
		got := envAppsUpdate{Repo: "nullstone-io/backend", Enable: true}.Automations(backend, current)

		assert.Equal(t, []api.UpdateWorkspaceAutomationInput{
			{BlockId: 1, IsEnabled: true, AutoDeploy: true},
			{BlockId: 2, IsEnabled: true, AutoDeploy: false},
		}, got)
	})
}

func TestDescribeTracking(t *testing.T) {
	assert.Equal(t, "branch main", describeTracking(types.PreviewApp{BranchName: strPtr("main")}))
	assert.Equal(t, "pull request 42", describeTracking(types.PreviewApp{PullRequestId: i64Ptr(42)}))
	assert.Equal(t, "tag", describeTracking(types.PreviewApp{TriggerType: "tag"}))
	assert.Equal(t, "-", describeTracking(types.PreviewApp{}))
}
