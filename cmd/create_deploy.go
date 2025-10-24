package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, info artifacts.VersionInfo) (*api.DeployCreateResult, error) {
	ctx := context.TODO()
	client := api.Client{Config: nsConfig}
	payload := api.DeployCreatePayload{
		FromSource:     false,
		Version:        info.EffectiveVersion,
		CommitSha:      info.CommitInfo.CommitSha,
		AutomationTool: detectAutomationTool(),
	}
	result, err := client.Deploys().Create(ctx, appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating deploy: %w", err)
	} else if result == nil {
		return nil, fmt.Errorf("unable to create deploy")
	}
	return result, nil
}

func detectAutomationTool() string {
	if os.Getenv("CIRCLECI") != "" {
		return api.AutomationToolCircleCI
	}
	if os.Getenv("GITHUB_ACTIONS") != "" {
		return api.AutomationToolGithubActions
	}
	if os.Getenv("GITLAB_CI") != "" {
		return api.AutomationToolGitlab
	}
	if os.Getenv("BITBUCKET_PIPELINES") != "" {
		return api.AutomationToolBitbucket
	}
	if os.Getenv("JENKINS_URL") != "" {
		return api.AutomationToolJenkins
	}
	if os.Getenv("TRAVIS") != "" {
		return api.AutomationToolTravis
	}
	if os.Getenv("TF_BUILD") != "" {
		// TF_BUILD is not referring to Terraform, it's legacy from the original system called "Team Foundation"
		return api.AutomationToolAzurePipelines
	}
	if os.Getenv("APPVEYOR") != "" {
		return api.AutomationToolAppveyor
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return api.AutomationToolTeamCity
	}
	if os.Getenv("CI_NAME") != "codeship" {
		return api.AutomationToolCodeship
	}
	if os.Getenv("SEMAPHORE") != "" {
		return api.AutomationToolSemaphore
	}
	return ""
}
