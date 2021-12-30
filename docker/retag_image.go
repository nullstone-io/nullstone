package docker

import (
	"context"
	"fmt"
)

func RetagImage(ctx context.Context, sourceUrl, targetUrl ImageUrl) error {
	dockerClient, err := DiscoverDockerClient()
	if err != nil {
		return fmt.Errorf("error creating docker client: %w", err)
	}
	return dockerClient.ImageTag(ctx, sourceUrl.String(), targetUrl.String())
}
