package aws_ecr

import (
	"context"
	"fmt"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"strings"
)

func PushImage(sourceUrl, targetUrl docker.ImageUrl, imagePusher aws.User, region string) error {
	// NOTE: We expect --version from the user which is used as the image tag for the pushed image
	if targetUrl.Tag == "" {
		return fmt.Errorf("no version was specified, version is required to push image")
	}
	if !strings.Contains(targetUrl.Registry, "ecr") &&
		!strings.Contains(targetUrl.Registry, "amazonaws.com") {
		return fmt.Errorf("this app only supports push to AWS ECR (image=%s)", targetUrl)
	}
	// NOTE: For now, we are assuming that the production docker image is hosted in ECR
	// This will likely need to be refactored to support pushing to other image registries
	if imagePusher.AccessKeyId == "" {
		return fmt.Errorf("cannot push without an authorized user, make sure 'image_pusher' output is not empty")
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	targetAuth, err := nsaws.GetEcrLoginAuth(ctx, imagePusher, region)
	if err != nil {
		return fmt.Errorf("error retrieving image registry credentials: %w", err)
	}

	logger.Printf("Retagging %s => %s\n", sourceUrl.String(), targetUrl.String())
	if err := docker.RetagImage(ctx, sourceUrl, targetUrl); err != nil {
		return fmt.Errorf("error retagging image: %w", err)
	}

	logger.Printf("Pushing %s\n", targetUrl.String())
	if err := docker.PushImage(ctx, targetUrl, targetAuth); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}
