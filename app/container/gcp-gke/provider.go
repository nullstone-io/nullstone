package gcp_gke

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
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

// Deploy takes the following steps to deploy an AWS Fargate service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("Not supported yet")
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, fmt.Errorf("Not supported yet")
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	reports := app.StatusDetailReports{}
	return reports, fmt.Errorf("Not supported yet")
}
