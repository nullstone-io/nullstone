package aws_lambda_zip

import (
	"context"
	"fmt"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
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
	Subplatform: "zip",
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

// Push will upload the versioned artifact to the source artifact bucket for the lambda
func (p Provider) Push(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	source := userConfig["source"]
	if source == "" {
		return fmt.Errorf("--source is required to upload artifact")
	}
	version := userConfig["version"]
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
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	// TODO: Add cancellation support so users can press Control+C to kill deploy
	ctx := context.TODO()

	logger.Printf("Deploying app %q\n", details.App.Name)
	version := userConfig["version"]
	if version == "" {
		return fmt.Errorf("--version is required to deploy app")
	}

	logger.Printf("Updating app version to %q\n", version)
	if err := app.CreateDeploy(nsConfig, details.App.StackId, details.App.Id, details.Env.Id, version); err != nil {
		return fmt.Errorf("error updating app version in nullstone: %w", err)
	}

	logger.Printf("Updating lambda to %q\n", version)
	if err := ic.UpdateLambdaVersion(ctx, version); err != nil {
		return fmt.Errorf("error updating lambda version: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil
}

func (p Provider) Exec(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("exec is not implemented for the lambda:zip provider yet")
}

func (p Provider) Ssh(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	return fmt.Errorf("ssh is not supported for the lambda:zip provider")
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, []ecstypes.ServiceEvent, error) {
	return app.StatusReport{}, nil, nil
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	return app.StatusDetailReports{}, nil
}
