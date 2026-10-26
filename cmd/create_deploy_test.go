package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

// ciEnvVars is every variable detectAutomationTool reads; each case starts with all of them
// blank so the developer's own shell (or the CI running these tests) cannot leak in.
var ciEnvVars = []string{"CIRCLECI", "GITHUB_ACTIONS", "GITLAB_CI", "BITBUCKET_PIPELINES", "JENKINS_URL",
	"TRAVIS", "TF_BUILD", "APPVEYOR", "TEAMCITY_VERSION", "CI_NAME", "SEMAPHORE"}

func TestDetectAutomationTool(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "no CI variables is a manual run", env: map[string]string{}, want: ""},
		{name: "github actions", env: map[string]string{"GITHUB_ACTIONS": "true"}, want: api.AutomationToolGithubActions},
		{name: "circleci", env: map[string]string{"CIRCLECI": "true"}, want: api.AutomationToolCircleCI},
		{name: "gitlab", env: map[string]string{"GITLAB_CI": "true"}, want: api.AutomationToolGitlab},
		{name: "azure pipelines", env: map[string]string{"TF_BUILD": "True"}, want: api.AutomationToolAzurePipelines},
		{name: "codeship", env: map[string]string{"CI_NAME": "codeship"}, want: api.AutomationToolCodeship},
		// Regression: the check used to be `!= "codeship"`, so every laptop run with no CI
		// variables at all was reported as a Codeship automation run.
		{name: "an unrelated CI_NAME is not codeship", env: map[string]string{"CI_NAME": "something-else"}, want: ""},
		{name: "semaphore", env: map[string]string{"SEMAPHORE": "true"}, want: api.AutomationToolSemaphore},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range ciEnvVars {
				t.Setenv(key, "")
			}
			for key, value := range test.env {
				t.Setenv(key, value)
			}
			assert.Equal(t, test.want, detectAutomationTool())
		})
	}
}
