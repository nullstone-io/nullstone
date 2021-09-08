package cmd

import "github.com/urfave/cli/v2"

var AppVersionFlag = &cli.StringFlag{
	Name: "version",
	Usage: `Push the artifact with this version.
       If not specified, will retrieve short sha from your latest commit.
       Specify '-' to ignore automatic version detection.
       app/container: If specified, will push the docker image with version as the image tag. Otherwise, uses source tag.
       app/serverless: This is required to upload the artifact.`,
}

func DetectAppVersion(c *cli.Context) string {
	version := c.String("version")
	switch version {
	case "":
		// If user does not specify a version, use HEAD commit sha
		if hash, err := getCurrentCommitSha(); err == nil {
			return hash[0:8]
		}
	case "-":
		// Ignore version in app command
		return ""
	}
	return version
}
