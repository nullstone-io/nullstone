package aws_s3

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
)

var _ app.Provider = Provider{}

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppStaticSite),
	Provider:    "aws",
	Platform:    "s3",
	Subplatform: "",
}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "s3"
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

func (p Provider) Push(logger *log.Logger, nsConfig api.Config, details app.Details, source, version string) error {
	ic, err := p.identify(logger, nsConfig, details)
	if err != nil {
		return err
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()
	if source == "" {
		return fmt.Errorf("no source specified, source artifact (directory or achive) is required to push")
	}
	if version == "" {
		return fmt.Errorf("no version specified, version is required to push")
	}

	filepaths, err := artifacts.WalkDir(source)
	if err != nil {
		return fmt.Errorf("error scanning source: %w", err)
	}

	logger.Printf("Uploading %s to s3 bucket %s...\n", source, ic.Outputs.BucketName)
	if err := ic.UploadArtifact(ctx, source, filepaths, version); err != nil {
		return fmt.Errorf("error uploading artifact: %w", err)
	}

	return nil
}

func (p Provider) Exec(ctx context.Context, logger *log.Logger, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("exec is not supported for the s3 provider")
}

func (p Provider) Ssh(ctx context.Context, logger *log.Logger, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	return fmt.Errorf("ssh is not supported for the s3 provider")
}

func (p Provider) Deploy(logger *log.Logger, nsConfig api.Config, details app.Details, version string) (*string, error) {
	ic, err := p.identify(logger, nsConfig, details)
	if err != nil {
		return nil, err
	}

	// TODO: Add cancellation support so users can press Control+C to kill deploy
	ctx := context.TODO()

	logger.Printf("Deploying app %q\n", details.App.Name)
	if version == "" {
		return nil, fmt.Errorf("no version specified, version is required to deploy")
	}

	logger.Printf("Updating CDN version to %q\n", version)
	if err := ic.UpdateCdnVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("error updating CDN version: %w", err)
	}

	logger.Println("Invalidating cache in CDNs")
	if err := ic.InvalidateCdnPaths(ctx, []string{"/*"}); err != nil {
		return nil, fmt.Errorf("error invalidating /*: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil, nil
}

func (p Provider) Status(logger *log.Logger, nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	// TODO: Implement me
	return app.StatusReport{}, fmt.Errorf("status is not supported for the s3 provider")
}

func (p Provider) DeploymentStatus(logger *log.Logger, nsConfig api.Config, deployReference string, details app.Details) (app.RolloutStatus, error) {
	return app.RolloutStatusUnknown, fmt.Errorf("deployment status is not supported for the s3 provider")
}

func (p Provider) StatusDetail(logger *log.Logger, nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	// TODO: Implement me
	return app.StatusDetailReports{}, fmt.Errorf("status detail is not supported for the s3 provider")
}
