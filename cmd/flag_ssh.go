package cmd

import "github.com/urfave/cli/v2"

var InstanceFlag = &cli.StringFlag{
	Name: "instance",
	Usage: `Select a specific instance to execute the command against.
		This allows the user to decide which instance to connect.
		This is optional and by default will connect to a random instance.
		This is only used for workspaces that use VMs (e.g. Elastic Beanstalk, EC2 Instances, GCP VMs, Azure VMs, etc.).`,
}
var TaskFlag = &cli.StringFlag{
	Name: "task",
	Usage: `Select a specific task to execute the command against.
		This is optional and by default will connect to a random task.
        This is only used by ECS and determines which task to connect.`,
}
var PodFlag = &cli.StringFlag{
	Name: "pod",
	Usage: `Select a pod to execute the command against.
        When specified, allows you to connect to a specific pod within a replica set.
        This is optional and will connect to a random pod by default.
        This is only used by Kubernetes clusters and determines which pod in the replica to connect.`,
}

var ContainerFlag = &cli.StringFlag{
	Name: "container",
	Usage: `Select a specific container within a task or pod.
        If using sidecars, this allows you to connect to other containers besides the primary application container.`,
}
