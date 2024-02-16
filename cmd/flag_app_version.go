package cmd

import (
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

var AppVersionFlag = &cli.StringFlag{
	Name: "version",
	Usage: `Provide a label for your deployment.
		If not provided, it will default to the commit sha of the repo for the current directory.`,
}

func DetectAppVersion(c *cli.Context) string {
	version := c.String("version")
	if version == "" {
		// If user does not specify a version, use HEAD commit sha
		if hash, err := vcs.GetCurrentCommitSha(); err == nil && len(hash) >= 7 {
			return hash[0:7]
		}
	}
	return version
}
