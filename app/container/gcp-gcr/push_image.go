package gcp_gcr

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/gcp"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/nsgcp"
	"log"
	"os"
	"strings"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

func PushImage(sourceUrl, targetUrl docker.ImageUrl, imagePusher gcp.ServiceAccount) error {
	// NOTE: We expect --version from the user which is used as the image tag for the pushed image
	if targetUrl.Tag == "" {
		return fmt.Errorf("no version was specified, version is required to push image")
	}
	if !strings.Contains(targetUrl.Registry, "gcr.io") {
		return fmt.Errorf("this app only supports push to GCP GCR (image=%s)", targetUrl)
	}
	// NOTE: For now, we are assuming that the production docker image is hosted in GCR
	// This will likely need to be refactored to support pushing to other image registries
	if imagePusher.PrivateKey == "" {
		return fmt.Errorf("cannot push without an authorized user, make sure 'image_pusher' output is not empty")
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	targetAuth, err := nsgcp.GetGcrLoginAuth(ctx, imagePusher, targetUrl.Registry)
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
