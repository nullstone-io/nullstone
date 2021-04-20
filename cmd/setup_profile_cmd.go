package cmd

import (
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

func SetupProfileCmd(c *cli.Context) (*config.Profile, api.Config, error) {
	profile, err := config.LoadProfile(GetProfile(c))
	if err != nil {
		return nil, api.Config{}, err
	}

	cfg := api.DefaultConfig()
	if profile.Address != "" {
		cfg.BaseAddress = profile.Address
	}
	if profile.ApiKey != "" {
		cfg.ApiKey = profile.ApiKey
	}
	cfg.ApiKey = config.CleanseApiKey(cfg.ApiKey)
	cfg.OrgName = GetOrg(c, *profile)
	if cfg.OrgName == "" {
		return profile, cfg, ErrMissingOrg
	}
	return profile, cfg, nil
}
