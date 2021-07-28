package gcp_gke

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
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
	return ""
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
	return fmt.Errorf("Not supported yet")
}

// Deploy takes the following steps to deploy a GCP GKE pod
//   Get pod
//   Change image tag
//   Update pod
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	pod, err := ic.GetPod()
	if err != nil {
		return fmt.Errorf("error retrieving pod: %w", err)
	}
	spec := pod.Spec

	logger.Printf("Deploying app %q\n", details.App.Name)
	version := userConfig["version"]
	if version != "" {
		logger.Printf("Updating app version to %q\n", version)
		if err := app.UpdateVersion(nsConfig, details.App.Id, details.Env.Name, version); err != nil {
			return fmt.Errorf("error updating app version in nullstone: %w", err)
		}

		logger.Printf("Updating image tag to %q\n", version)
		if spec, err = ic.ReplacePodSpecImageTag(pod.Spec, version); err != nil {
			return fmt.Errorf("error updating pod spec with new image tag: %w", err)
		}
	}

	pod.Spec = spec
	if _, err := ic.UpdatePod(pod); err != nil {
		return fmt.Errorf("error deploying service: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, fmt.Errorf("Not supported yet")
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	reports := app.StatusDetailReports{}
	return reports, fmt.Errorf("Not supported yet")
}
