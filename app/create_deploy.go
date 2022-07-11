package app

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
)

func CreateDeploy(nsConfig api.Config, stackId, appId, envId int64, version string) error {
	if version == "" {
		return fmt.Errorf("no version specified, version is required to create a deploy")
	}

	client := api.Client{Config: nsConfig}
	result, err := client.Deploys().Create(stackId, appId, envId, version)
	if err != nil {
		return fmt.Errorf("error updating app version: %w", err)
	} else if result == nil {
		return fmt.Errorf("could not find application environment")
	}

	log.Println("Deployment created")
	return nil
}
