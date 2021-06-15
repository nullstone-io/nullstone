package cmd

import "github.com/urfave/cli/v2"

var AppSourceFlag = &cli.StringFlag{
	Name: "source",
	Usage: `The source artifact to push.
       app/container: This is the docker image to push. This follows the same syntax as 'docker push NAME[:TAG]'.
       app/serverless: This is a .zip archive to push.`,
	Required: true,
}
