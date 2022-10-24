package cmd

import "github.com/urfave/cli/v2"

var AppSourceFlag = &cli.StringFlag{
	Name: "source",
	Usage: `The source artifact to push that contains your application's build.
		For a container, specify the name of the docker image to push. This follows the same syntax as 'docker push NAME[:TAG]'.
		For a serverless zip application, specify the .zip archive to push.`,
	Required: true,
}
