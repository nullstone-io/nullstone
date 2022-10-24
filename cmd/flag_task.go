package cmd

import "github.com/urfave/cli/v2"

var TaskFlag = &cli.StringFlag{
	Name: "task",
	Usage: `Select a specific task/replica to execute the command against.
		This is optional and by default will connect to a random task/replica.
       	If using Kubernetes, this will select which replica of the pod to connect.
       	If using ECS, this will select which task of the service to connect.`,
}
