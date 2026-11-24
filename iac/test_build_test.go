package iac

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nullstone-io/iac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

const (
	buildTestRepoUrl  = "https://github.com/acme/repo"
	buildTestRepoName = "acme/repo"

	buildTestConfigYml = `version: "0.1"
apps:
  api:
    module: nullstone/aws-fargate-service
    build:
      builder: github-actions
      github-actions:
        workflow_filename: .github/workflows/api.yml
        workflow_inputs:
          target: api
`
	buildTestConfigYmlNoBuild = `version: "0.1"
apps:
  api:
    module: nullstone/aws-fargate-service
`
	buildTestProdYml = `version: "0.1"
apps:
  api:
    build:
      github-actions:
        initiate_workflow: false
`
	// buildTestProdNullstoneYml names a setting of the nullstone builder and no builder
	buildTestProdNullstoneYml = `version: "0.1"
apps:
  api:
    build:
      nullstone:
        dockerfile: api/Dockerfile.prod
`
)

// buildTestApi serves the three reads testBuildConfig makes: the workspace build config, the block template
// and the module. A nil current means the workspace has no build config (it does not exist).
func buildTestApi(t *testing.T, current *types.WorkspaceBuildConfig, template types.WorkspaceTemplate) *api.Client {
	t.Helper()
	send := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/stacks/1/blocks/2/envs/3/build_config", func(w http.ResponseWriter, r *http.Request) {
		if current == nil {
			http.NotFound(w, r)
			return
		}
		send(w, current)
	})
	mux.HandleFunc("/orgs/acme/stacks/1/blocks/2/workspace_template", func(w http.ResponseWriter, r *http.Request) {
		send(w, template)
	})
	mux.HandleFunc("/orgs/nullstone/modules/aws-fargate-service", func(w http.ResponseWriter, r *http.Request) {
		send(w, types.Module{OrgName: "nullstone", Name: "aws-fargate-service", Category: types.CategoryApp, Subcategory: types.SubcategoryAppContainer})
	})
	mux.HandleFunc("/orgs/nullstone/modules/aws-lambda-zip", func(w http.ResponseWriter, r *http.Request) {
		send(w, types.Module{OrgName: "nullstone", Name: "aws-lambda-zip", Category: types.CategoryApp, Subcategory: types.SubcategoryAppServerless, Subplatform: "zip"})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return &api.Client{Config: api.Config{BaseAddress: server.URL, OrgName: "acme"}}
}

func runTestBuildConfig(t *testing.T, apiClient *api.Client, blockType string, envName string, files map[string]string) string {
	t.Helper()
	got, err := tryTestBuildConfig(t, apiClient, blockType, envName, files)
	require.NoError(t, err)
	return got
}

func tryTestBuildConfig(t *testing.T, apiClient *api.Client, blockType string, envName string, files map[string]string) (string, error) {
	t.Helper()
	pmr, err := iac.ParseMap(buildTestRepoUrl, buildTestRepoName, files)
	require.NoError(t, err)
	stack := types.Stack{IdModel: types.IdModel{Id: 1}, OrgName: "acme", Name: "primary"}
	block := types.Block{IdModel: types.IdModel{Id: 2}, OrgName: "acme", StackId: 1, Type: blockType, Name: "api", ModuleSource: "nullstone/aws-fargate-service"}
	env := types.Environment{IdModel: types.IdModel{Id: 3}, OrgName: "acme", StackId: 1, Type: types.EnvTypePipeline, Name: envName}

	var buf bytes.Buffer
	err = testBuildConfig(context.Background(), apiClient, &buf, stack, block, env, pmr)
	return ansiRe.ReplaceAllString(buf.String(), ""), err
}

func TestTestBuildConfig(t *testing.T) {
	app := string(types.BlockTypeApplication)
	moduleDefault := types.DefaultBuildConfig(nil)
	template := types.WorkspaceTemplate{Config: types.WorkspaceTemplateConfig{Module: "nullstone/aws-fargate-service"}}
	fromThisRepo := &types.BuildConfigSource{Repo: buildTestRepoName, RepoUrl: buildTestRepoUrl, File: "config.yml"}

	t.Run("config.yml and the overrides file of the env are laid over the module default", func(t *testing.T) {
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: moduleDefault}, template)

		got := runTestBuildConfig(t, apiClient, app, "prod", map[string]string{"config.yml": buildTestConfigYml, "prod.yml": buildTestProdYml})

		want := "    api => 4 build config changes (applied to the next build)\n" +
			"        ~ build.builder: nullstone => github-actions\n" +
			"        ~ build.github-actions.workflow_filename: .github/workflows/build.yml => .github/workflows/api.yml\n" +
			"        ~ build.github-actions.initiate_workflow: true => false\n" +
			"        ~ build.github-actions.workflow_inputs: {} => {target: api}\n"
		assert.Equal(t, want, got)
	})

	t.Run("an env without an overrides file takes config.yml", func(t *testing.T) {
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: moduleDefault}, template)

		got := runTestBuildConfig(t, apiClient, app, "dev", map[string]string{"config.yml": buildTestConfigYml, "prod.yml": buildTestProdYml})

		assert.Contains(t, got, "api => 3 build config changes")
		assert.NotContains(t, got, "initiate_workflow")
	})

	t.Run("no file declares build: a build config set outside IaC is left alone", func(t *testing.T) {
		edited := moduleDefault
		edited.Dockerfile = "ui/Dockerfile"
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: edited}, template)

		got := runTestBuildConfig(t, apiClient, app, "dev", map[string]string{"config.yml": buildTestConfigYmlNoBuild})

		assert.Empty(t, got)
	})

	t.Run("build removed from every file: a build config IaC wrote keeps its settings", func(t *testing.T) {
		written := moduleDefault
		written.Dockerfile = "api/Dockerfile"
		templateBuild := moduleDefault
		templateBuild.Dockerfile = "template/Dockerfile"
		withBuild := template
		withBuild.Config.Build = &templateBuild
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: written, Source: fromThisRepo}, withBuild)

		got := runTestBuildConfig(t, apiClient, app, "dev", map[string]string{"config.yml": buildTestConfigYmlNoBuild})

		assert.Empty(t, got)
	})

	t.Run("the package mode comes from the module, not from the block template's build", func(t *testing.T) {
		// The block template's build was saved for a static site; the module is a container
		templateBuild := moduleDefault
		templateBuild.PackageMode = types.BuildPackageModeSiteAssets
		withBuild := template
		withBuild.Config.Build = &templateBuild
		current := moduleDefault
		current.PackageMode = types.BuildPackageModeSiteAssets
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: current}, withBuild)

		got := runTestBuildConfig(t, apiClient, app, "prod", map[string]string{"config.yml": buildTestConfigYmlNoBuild, "prod.yml": buildTestProdNullstoneYml})

		want := "    api => 2 build config changes (applied to the next build)\n" +
			"        ~ build.package_mode: site-assets => dockerfile\n" +
			"        ~ build.nullstone.dockerfile: Dockerfile => api/Dockerfile.prod\n"
		assert.Equal(t, want, got)
	})

	t.Run("settings the builder of the workspace does not use are an error", func(t *testing.T) {
		// No file names a builder, so prod.yml is checked against the build config it is laid over
		manual := moduleDefault
		manual.Builder = types.BuilderManual
		withBuild := template
		withBuild.Config.Build = &manual
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: manual}, withBuild)

		got, err := tryTestBuildConfig(t, apiClient, app, "prod", map[string]string{"config.yml": buildTestConfigYmlNoBuild, "prod.yml": buildTestProdNullstoneYml})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "apps.api.build.nullstone => Settings of the nullstone builder are not used, the builder is (manual)")
		assert.Empty(t, got)
	})

	t.Run("the nullstone builder on a zip module is an error", func(t *testing.T) {
		// The app is a zip module and no file names a builder, so it is laid over the default: nullstone
		zipTemplate := types.WorkspaceTemplate{Config: types.WorkspaceTemplateConfig{Module: "nullstone/aws-lambda-zip"}}
		zipDefault := types.DefaultBuildConfig(&types.Module{Subplatform: "zip"})
		zipYml := `version: "0.1"
apps:
  api:
    build:
      nullstone:
        publish_dir: ./out
`
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: zipDefault}, zipTemplate)

		got, err := tryTestBuildConfig(t, apiClient, app, "prod", map[string]string{"prod.yml": zipYml})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "apps.api.build.builder => The nullstone builder cannot build the package mode (zip) of the app's module")
		assert.Empty(t, got)
	})

	t.Run("another builder on a zip module is reported with the package mode of the module", func(t *testing.T) {
		zipTemplate := types.WorkspaceTemplate{Config: types.WorkspaceTemplateConfig{Module: "nullstone/aws-lambda-zip"}}
		zipDefault := types.DefaultBuildConfig(&types.Module{Subplatform: "zip"})
		zipYml := `version: "0.1"
apps:
  api:
    build:
      builder: github-actions
`
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: zipDefault}, zipTemplate)

		got := runTestBuildConfig(t, apiClient, app, "prod", map[string]string{"prod.yml": zipYml})

		want := "    api => 1 build config change (applied to the next build)\n" +
			"        ~ build.builder: nullstone => github-actions\n"
		assert.Equal(t, want, got)
	})

	t.Run("settings the module of the workspace does not use are an error", func(t *testing.T) {
		siteYml := `version: "0.1"
apps:
  api:
    build:
      nullstone:
        build_cmd: yarn build
`
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: moduleDefault}, template)

		got, err := tryTestBuildConfig(t, apiClient, app, "prod", map[string]string{"prod.yml": siteYml})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "apps.api.build.nullstone.build_cmd => Setting is not used by the package mode (dockerfile) of the app's module")
		assert.Empty(t, got)
	})

	t.Run("nothing to report when the build config already matches", func(t *testing.T) {
		matching := moduleDefault
		matching.Builder = "github-actions"
		matching.WorkflowFilename = ".github/workflows/api.yml"
		matching.WorkflowInputs = map[string]string{"target": "api"}
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: matching, Source: fromThisRepo}, template)

		got := runTestBuildConfig(t, apiClient, app, "dev", map[string]string{"config.yml": buildTestConfigYml})

		assert.Empty(t, got)
	})

	t.Run("a workspace that does not exist is skipped", func(t *testing.T) {
		apiClient := buildTestApi(t, nil, template)

		got := runTestBuildConfig(t, apiClient, app, "dev", map[string]string{"config.yml": buildTestConfigYml})

		assert.Empty(t, got)
	})

	t.Run("a block that is not an app is skipped", func(t *testing.T) {
		apiClient := buildTestApi(t, &types.WorkspaceBuildConfig{Config: moduleDefault}, template)

		got := runTestBuildConfig(t, apiClient, string(types.BlockTypeDatastore), "dev", map[string]string{"config.yml": buildTestConfigYml})

		assert.Empty(t, got)
	})
}

func TestBuildValToString(t *testing.T) {
	assert.Equal(t, "(empty)", buildValToString(""))
	assert.Equal(t, "Dockerfile", buildValToString("Dockerfile"))
	assert.Equal(t, "true", buildValToString(true))
	assert.Equal(t, "{}", buildValToString(map[string]string{}))
	assert.Equal(t, "{a: 1, b: 2}", buildValToString(map[string]string{"b": "2", "a": "1"}))
}
