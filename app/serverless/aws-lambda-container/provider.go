package aws_lambda_container

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app/container/aws-ecr"
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
	Subcategory: string(types.SubcategoryAppServerless),
	Provider:    "aws",
	Platform:    "lambda",
	Subplatform: "container",
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

// Push will upload the versioned artifact to the ECR repository for the lambda
func (p Provider) Push(nsConfig api.Config, details app.Details, source, version string) error {
	return (aws_ecr.Provider{}).Push(nsConfig, details, source, version)
}

// Deploy takes the following steps to deploy an AWS Lambda service
//   Update app version in nullstone
//   Update function code to use just-uploaded image
func (p Provider) Deploy(nsConfig api.Config, details app.Details, version string) (*string, error) {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return nil, err
	}

	// TODO: Add cancellation support so users can press Control+C to kill deploy
	ctx := context.TODO()

	logger.Printf("Deploying app %q\n", details.App.Name)
	if version == "" {
		return nil, fmt.Errorf("--version is required to deploy app")
	}

	logger.Printf("Updating lambda to %q\n", version)
	if err := ic.UpdateLambdaVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("error updating lambda version: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil, nil
}

func (p Provider) Exec(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("exec is not implemented for the lambda:container provider yet")
}

func (p Provider) Ssh(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	return fmt.Errorf("ssh is not supported for the lambda:container provider")
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, fmt.Errorf("status is not supported for the lambda:container provider")
}

func (p Provider) DeploymentStatus(deployReference string, nsConfig api.Config, details app.Details) (app.StatusReport, []app.ServiceEvent, error) {
	return app.StatusReport{}, nil, fmt.Errorf("deployment status is not supported for the lambda:container provider")
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	return app.StatusDetailReports{}, fmt.Errorf("status detail is not supported for the lambda:container provider")
}
