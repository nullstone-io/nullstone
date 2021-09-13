package cmd

import "github.com/urfave/cli/v2"

var AppVersionFlag = &cli.StringFlag{
	Name: "version",
	Usage: `Push/Deploy the artifact with this version.
       If not specified, will retrieve short sha from your latest commit.
       app/container: If specified, will push the docker image with version as the image tag. Otherwise, uses source tag.
       app/serverless: This is required to upload the artifact.`,
}

func DetectAppVersion(c *cli.Context) string {
	version := c.String("version")
	if version == "" {
		// If user does not specify a version, use HEAD commit sha
		if hash, err := getCurrentCommitSha(); err == nil {
			return hash[0:8]
		}
	}
	return version
}
