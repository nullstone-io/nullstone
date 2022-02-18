package aws_s3

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
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
	return "s3"
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

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()
	source := userConfig["source"]
	if source == "" {
		return fmt.Errorf("no source specified, source artifact (directory or achive) is required to push")
	}
	version := userConfig["version"]
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

func (p Provider) Exec(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	return fmt.Errorf("exec is not supported for the aws-s3 provider")
}

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
		return fmt.Errorf("no version specified, version is required to deploy")
	}

	logger.Printf("Updating app version to %q\n", version)
	if err := app.UpdateVersion(nsConfig, details.App.Id, details.Env.Name, version); err != nil {
		return fmt.Errorf("error updating app version in nullstone: %w", err)
	}

	logger.Printf("Updating CDN version to %q\n", version)
	if err := ic.UpdateCdnVersion(ctx, version); err != nil {
		return fmt.Errorf("error updating CDN version: %w", err)
	}

	logger.Println("Invalidating sitemap.xml cache in CDNs")
	if err := ic.InvalidateCdnPaths(ctx, []string{"/sitemap.xml"}); err != nil {
		return fmt.Errorf("error invalidating sitemap.xml: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	// TODO: Implement me
	return app.StatusReport{}, nil
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	// TODO: Implement me
	return app.StatusDetailReports{}, nil
}
