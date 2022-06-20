package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app/container/aws-ecr"
	"gopkg.in/nullstone-io/nullstone.v0/app/container/aws-ecs-fargate"
	"gopkg.in/nullstone-io/nullstone.v0/app/server/aws-ec2"
	"gopkg.in/nullstone-io/nullstone.v0/app/serverless/aws-lambda-container"
	"gopkg.in/nullstone-io/nullstone.v0/app/serverless/aws-lambda-zip"
	"gopkg.in/nullstone-io/nullstone.v0/app/static-site/aws-s3"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs/aws/cloudwatch"
	"gopkg.in/nullstone-io/nullstone.v0/cmd"
	"os"
	"sort"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	appProviders := app.Providers{
		aws_ecr.ModuleContractName:              aws_ecr.Provider{},
		aws_ecs_fargate.ModuleContractName:      aws_ecs_fargate.Provider{},
		aws_s3.ModuleContractName:               aws_s3.Provider{},
		aws_lambda_zip.ModuleContractName:       aws_lambda_zip.Provider{},
		aws_lambda_container.ModuleContractName: aws_lambda_container.Provider{},
		aws_ec2.ModuleContractName:              aws_ec2.Provider{},
	}
	logProviders := app_logs.Providers{
		"cloudwatch": cloudwatch.Provider{},
	}

	cliApp := &cli.App{
		Version:              version,
		EnableBashCompletion: true,
		Metadata: map[string]interface{}{
			"commit":  commit,
			"date":    date,
			"builtBy": builtBy,
		},
		Flags: []cli.Flag{
			cmd.ProfileFlag,
			cmd.OrgFlag,
		},
		Commands: []*cli.Command{
			{
				Name: "version",
				Action: func(c *cli.Context) error {
					cli.ShowVersion(c)
					return nil
				},
			},
			cmd.Configure,
			cmd.SetOrg,
			cmd.Stacks,
			cmd.Envs,
			cmd.Apps,
			cmd.Blocks,
			cmd.Modules,
			cmd.Workspaces,
			cmd.Up(),
			cmd.Outputs(),
			cmd.Push(appProviders),
			cmd.Deploy(appProviders),
			cmd.Launch(appProviders, logProviders),
			cmd.Logs(appProviders, logProviders),
			cmd.Status(appProviders),
			cmd.Exec(appProviders),
			cmd.Ssh(appProviders),
		},
	}
	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
