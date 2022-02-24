package app

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func UpdateVersion(nsConfig api.Config, stackId, appId int64, envName, version string) error {
	client := api.Client{Config: nsConfig}
	result, err := client.AppEnvs().Update(stackId, appId, envName, version)
	if err != nil {
		return fmt.Errorf("error updating app version: %w", err)
	} else if result == nil {
		return fmt.Errorf("could not find application environment")
	}
	return nil
}
