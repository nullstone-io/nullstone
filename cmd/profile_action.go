package cmd

import (
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type ProfileFn func(cfg api.Config) error

func ProfileAction(c *cli.Context, fn ProfileFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}
	return fn(cfg)
}
