package cmd

import "github.com/urfave/cli/v2"

var AppVersionFlag = &cli.StringFlag{
	Name: "version",
	Usage: `Push the artifact with this version.
       app/container: If specified, will push the docker image with version as the image tag. Otherwise, uses source tag.
       app/serverless: This is required to upload the artifact.`,
}
