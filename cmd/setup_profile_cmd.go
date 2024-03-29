package cmd

import (
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/auth"
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
		cfg.AccessTokenSource = auth.RawAccessTokenSource{AccessToken: profile.ApiKey}
	}
	cfg.OrgName = GetOrg(c, *profile)
	if cfg.OrgName == "" {
		return profile, cfg, ErrMissingOrg
	}

	if rats, ok := cfg.AccessTokenSource.(auth.RawAccessTokenSource); ok {
		cfg.AccessTokenSource = auth.RawAccessTokenSource{AccessToken: config.CleanseApiKey(rats.AccessToken)}
	}

	return profile, cfg, nil
}
