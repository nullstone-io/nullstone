package cmd

import "github.com/urfave/cli/v2"

var AppVersionFlag = &cli.StringFlag{
	Name: "version",
	Usage: `Provide a label for your deployment.
		If not provided, it will default to the commit sha of the repo for the current directory.`,
}

func DetectAppVersion(c *cli.Context) string {
	version := c.String("version")
	if version == "" {
		// If user does not specify a version, use HEAD commit sha
		if hash, err := getCurrentCommitSha(); err == nil && len(hash) >= 8 {
			return hash[0:8]
		}
	}
	return version
}
