package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
)

const (
	AutomationToolCircleCI       = "circleci"
	AutomationToolGithubActions  = "github-actions"
	AutomationToolGitlab         = "gitlab"
	AutomationToolBitbucket      = "bitbucket"
	AutomationToolJenkins        = "jenkins"
	AutomationToolTravis         = "travis"
	AutomationToolAzurePipelines = "azure-pipeline"
	AutomationToolAppveyor       = "appveyor"
	AutomationToolTeamCity       = "team-city"
	AutomationToolCodeship       = "codeship"
	AutomationToolSemaphore      = "semaphore"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, commitSha, version string) (*api.DeployCreateResult, error) {
	ctx := context.TODO()
	client := api.Client{Config: nsConfig}
	payload := api.DeployCreatePayload{
		FromSource:     false,
		Version:        version,
		CommitSha:      commitSha,
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
		return AutomationToolCircleCI
	}
	if os.Getenv("GITHUB_ACTIONS") != "" {
		return AutomationToolGithubActions
	}
	if os.Getenv("GITLAB_CI") != "" {
		return AutomationToolGitlab
	}
	if os.Getenv("BITBUCKET_PIPELINES") != "" {
		return AutomationToolBitbucket
	}
	if os.Getenv("JENKINS_URL") != "" {
		return AutomationToolJenkins
	}
	if os.Getenv("TRAVIS") != "" {
		return AutomationToolTravis
	}
	if os.Getenv("TF_BUILD") != "" {
		// TF_BUILD is not referring to Terraform, it's legacy from the original system called "Team Foundation"
		return AutomationToolAzurePipelines
	}
	if os.Getenv("APPVEYOR") != "" {
		return AutomationToolAppveyor
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return AutomationToolTeamCity
	}
	if os.Getenv("CI_NAME") != "codeship" {
		return AutomationToolCodeship
	}
	if os.Getenv("SEMAPHORE") != "" {
		return AutomationToolSemaphore
	}
	return ""
}
