package aws_ec2

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var _ app.Provider = Provider{}

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppServer),
	Provider:    "aws",
	Platform:    "ec2",
	Subplatform: "",
}

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

func (p Provider) Push(nsConfig api.Config, details app.Details, source, version string) error {
	return fmt.Errorf("push is not supported for the aws-ec2 provider")
}

func (p Provider) Deploy(nsConfig api.Config, details app.Details, version string) (*string, error) {
	return nil, fmt.Errorf("deploy is not supported for the aws-ec2 provider")
}

func (p Provider) Exec(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	return ic.ExecCommand(ctx, userConfig["cmd"], nil)
}

func (p Provider) Ssh(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	var parameters map[string][]string
	if val, ok := userConfig["forwards"].([]config.PortForward); ok {
		if parameters, err = ssm.SessionParametersFromPortForwards(val); err != nil {
			return err
		}
	}

	return ic.ExecCommand(ctx, "/bin/sh", parameters)
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, fmt.Errorf("status is not supported for the ec2 provider")
}

func (p Provider) DeploymentStatus(deployReference string, nsConfig api.Config, details app.Details) (app.StatusReport, []string, error) {
	return app.StatusReport{}, nil, fmt.Errorf("deployment status is not supported for the ec2 provider")
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	return app.StatusDetailReports{}, fmt.Errorf("status detail is not supported for the ec2 provider")
}
