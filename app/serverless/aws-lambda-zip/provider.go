package aws_lambda_zip

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
)

var _ app.Provider = Provider{}

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppServerless),
	Provider:    "aws",
	Platform:    "lambda",
	Subplatform: "zip",
}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "cloudwatch"
}

func (p Provider) identify(logger *log.Logger, nsConfig api.Config, details app.Details) (*InfraConfig, error) {
	logger.Printf("Identifying infrastructure for app %q\n", details.App.Name)
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)
	return ic, nil
}

// Push will upload the versioned artifact to the source artifact bucket for the lambda
func (p Provider) Push(logger *log.Logger, nsConfig api.Config, details app.Details, source, version string) error {
	ic, err := p.identify(logger, nsConfig, details)
	if err != nil {
		return err
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	if source == "" {
		return fmt.Errorf("--source is required to upload artifact")
	}
	if version == "" {
		return fmt.Errorf("--version is required to upload artifact")
	}

	file, err := os.Open(source)
	if os.IsNotExist(err) {
		return fmt.Errorf("source file %q does not exist", source)
	} else if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer file.Close()

	logger.Printf("Uploading %s to artifacts bucket\n", ic.Outputs.ArtifactsKey(version))
	if err := ic.UploadArtifact(ctx, file, version); err != nil {
		return fmt.Errorf("error uploading artifact: %w", err)
	}

	logger.Printf("Upload complete")

	return nil
}

// Deploy takes the following steps to deploy an AWS Lambda service
//   Update app version in nullstone
//   Update function code to use just-uploaded archive
func (p Provider) Deploy(logger *log.Logger, nsConfig api.Config, details app.Details, version string) (*string, error) {
	ic, err := p.identify(logger, nsConfig, details)
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

func (p Provider) Exec(ctx context.Context, logger *log.Logger, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("exec is not implemented for the lambda:zip provider yet")
}

func (p Provider) Ssh(ctx context.Context, logger *log.Logger, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	return fmt.Errorf("ssh is not supported for the lambda:zip provider")
}

func (p Provider) Status(logger *log.Logger, nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	return app.StatusReport{}, fmt.Errorf("status is not supported for the lambda:zip provider")
}

func (p Provider) DeploymentStatus(logger *log.Logger, nsConfig api.Config, deployReference string, details app.Details) (app.RolloutStatus, error) {
	return app.RolloutStatusUnknown, fmt.Errorf("deployment status is not supported for the lambda:zip provider")
}

func (p Provider) StatusDetail(logger *log.Logger, nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	return app.StatusDetailReports{}, fmt.Errorf("status detail is not supported for the lambda:zip provider")
}
