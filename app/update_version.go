package app

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func UpdateVersion(nsConfig api.Config, appId int64, envName, version string) error {
	client := api.Client{Config: nsConfig}
	_, err := client.AppEnvs().Update(appId, envName, version)
	if err != nil {
		return fmt.Errorf("error updating app version: %w", err)
	}
	return nil
}
