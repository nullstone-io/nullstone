package aws_ecr

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "cloudwatch"
}

func (p Provider) identify(nsConfig api.Config, details app.Details) (*InfraConfig, error) {
	logger.Printf("Identifying infrastructure for app %q\n", details.App.Name)
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)
	return ic, nil
}

func (p Provider) Push(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	sourceUrl := docker.ParseImageUrl(userConfig["source"])
	targetUrl := ic.Outputs.ImageRepoUrl
	if targetUrl.String() == "" {
		return fmt.Errorf("cannot push if 'image_repo_url' module output is missing")
	}
	targetUrl.Tag = userConfig["version"]

	return PushImage(sourceUrl, targetUrl, ic.Outputs.ImagePusher, ic.Outputs.Region)
}

// Deploy updates the app version
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	version := userConfig["version"]
	if version != "" {
		logger.Printf("Updating app version to %q\n", version)
		if err := app.UpdateVersion(nsConfig, details.App.Id, details.Env.Name, version); err != nil {
			return fmt.Errorf("error updating app version in nullstone: %w", err)
		}
	}
	return nil
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, nil
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	return app.StatusDetailReports{}, nil
}
