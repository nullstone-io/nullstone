package docker

import (
	"context"
	"fmt"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

func PushImage(ctx context.Context, targetUrl ImageUrl, targetAuth types.AuthConfig) error {
	encodedAuth, err := EncodeAuthToBase64(targetAuth)
	if err != nil {
		return fmt.Errorf("error encoding remote auth configuration: %w", err)
	}
	options := types.ImagePushOptions{
		All:          false,
		RegistryAuth: encodedAuth,
	}

	dockerClient, err := client.NewClientWithOpts()
	if err != nil {
		return fmt.Errorf("error creating docker client: %w", err)
	}
	responseBody, err := dockerClient.ImagePush(ctx, targetUrl.String(), options)
	if err != nil {
		return err
	}

	_, stdout, _ := term.StdStreams()
	return jsonmessage.DisplayJSONMessagesToStream(responseBody, streams.NewOut(stdout), nil)
}
