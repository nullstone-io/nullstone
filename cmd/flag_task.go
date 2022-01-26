package cmd

import "github.com/urfave/cli/v2"

var TaskFlag = &cli.StringFlag{
	Name: "task",
	Usage: `Optionally, specify the task/replica to execute the command against.
If not specified, this will connect to a random task/replica.
If using Kubernetes, this will select which replica of the pod to connect.
If using ECS, this will select which task of the service to connect.`,
}
