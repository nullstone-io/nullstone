package aws_lambda

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

var (
	logger = log.New(os.Stderr, "", 0)
)

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "cloudwatch"
}

func (p Provider) identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (*InfraConfig, error) {
	logger.Printf("Identifying infrastructure for app %q\n", app.Name)
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(workspace, &ic.Outputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)
	return ic, nil
}

// Push will upload the versioned artifact to the source artifact bucket for the lambda
func (p Provider) Push(nsConfig api.Config, application *types.Application, env *types.Environment, workspace *types.Workspace, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, application, workspace)
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
func (p Provider) Deploy(nsConfig api.Config, application *types.Application, env *types.Environment, workspace *types.Workspace, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, application, workspace)
	if err != nil {
		return err
	}

	// TODO: Add cancellation support so users can press Control+C to kill deploy
	ctx := context.TODO()

	version := userConfig["version"]
	if version == "" {
		return fmt.Errorf("--version is required to upload artifact")
	}

	logger.Printf("Deploying app %q\n", application.Name)

	logger.Printf("Updating app version to %q\n", version)
	if err := app.UpdateVersion(nsConfig, application.Id, env.Name, version); err != nil {
		return fmt.Errorf("error updating app version in nullstone: %w", err)
	}

	logger.Printf("Updating lambda to %q\n", version)
	if err := ic.UpdateLambdaVersion(ctx, version); err != nil {
		return fmt.Errorf("error updating lambda version: %w", err)
	}

	logger.Printf("Deployed app %q\n", application.Name)
	return nil
}
