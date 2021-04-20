package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
)

func DiscoverDockerClient() (client.APIClient, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}
	opts := &flags.ClientOptions{
		Common: &flags.CommonOptions{},
	}
	if err := dockerCli.Initialize(opts); err != nil {
		return nil, err
	}
	return dockerCli.Client(), nil
}
