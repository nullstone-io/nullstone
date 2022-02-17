package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	aws_ecr "gopkg.in/nullstone-io/nullstone.v0/app/container/aws-ecr"
	"gopkg.in/nullstone-io/nullstone.v0/app/container/aws-fargate"
	aws_ec2 "gopkg.in/nullstone-io/nullstone.v0/app/server/aws-ec2"
	aws_lambda "gopkg.in/nullstone-io/nullstone.v0/app/serverless/aws-lambda"
	aws_s3 "gopkg.in/nullstone-io/nullstone.v0/app/static-site/aws-s3"
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
		types.CategoryAppContainer: {
			"service/aws-fargate": aws_fargate.Provider{},
			"service/aws-ecr":     aws_ecr.Provider{},
		},
		types.CategoryAppStaticSite: {
			"site/aws-s3": aws_s3.Provider{},
		},
		types.CategoryAppServerless: {
			"service/aws-lambda": aws_lambda.Provider{},
		},
		types.CategoryAppServer: {
			"server/aws-ec2": aws_ec2.Provider{},
		},
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
			cmd.Apps,
			cmd.Stacks,
			cmd.Envs,
			cmd.Outputs(),
			cmd.Push(appProviders),
			cmd.Deploy(appProviders),
			cmd.Launch(appProviders, logProviders),
			cmd.Logs(appProviders, logProviders),
			cmd.Status(appProviders),
			cmd.Exec(appProviders),
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
