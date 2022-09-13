package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
	"os"
	"path"
	"strings"
)

var (
	moduleManifestFilename = path.Join(".nullstone", "module.yml")
	includeFlag            = &cli.StringSliceFlag{
		Name: "include",
		Usage: `Specify additional file patterns to package.
By default, this command includes *.tf, *.tf.tmpl, and README.md.
Use this flag to package additional modules and files needed for applies.
This supports file globbing detailed at https://pkg.go.dev/path/filepath#Glob`,
	}
)

var Modules = &cli.Command{
	Name:      "modules",
	Usage:     "View and modify modules",
	UsageText: "nullstone modules [subcommand]",
	Subcommands: []*cli.Command{
		ModulesGenerate,
		ModulesRegister,
		ModulesPublish,
		ModulesPackage,
	},
}

var ModulesGenerate = &cli.Command{
	Name:      "generate",
	Usage:     "Generate new module manifest (and optionally register)",
	UsageText: "nullstone modules generate [--register]",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "register"},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			existing, _ := modules.ManifestFromFile(moduleManifestFilename)
			survey := &moduleSurvey{}
			manifest, err := survey.Ask(cfg, existing)
			if err != nil {
				return err
			}
			if err := manifest.WriteManifestToFile(moduleManifestFilename); err != nil {
				return err
			}
			fmt.Printf("generated module manifest file to %s\n", moduleManifestFilename)

			if err := modules.Generate(manifest); err != nil {
				return err
			}
			fmt.Printf("generated base Terraform\n")

			if c.IsSet("register") {
				module, err := modules.Register(cfg, manifest)
				if err != nil {
					return err
				}
				fmt.Printf("registered %s/%s\n", module.OrgName, module.Name)
			}
			return nil
		})
	},
}

var ModulesRegister = &cli.Command{
	Name:      "register",
	Usage:     "Register module from .nullstone/module.yml",
	UsageText: "nullstone modules register",
	Flags:     []cli.Flag{},
	Aliases:   []string{"new"},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			manifest, err := modules.ManifestFromFile(moduleManifestFilename)
			if err != nil {
				return err
			}

			module, err := modules.Register(cfg, manifest)
			if err != nil {
				return err
			}
			fmt.Printf("registered %s/%s\n", module.OrgName, module.Name)
			return nil
		})
	},
}

var ModulesPublish = &cli.Command{
	Name:      "publish",
	Usage:     "Package and publish new version of a module",
	UsageText: "nullstone modules publish --version=<version>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage: `Specify a semver version for the module.
'next-patch': Uses a version that bumps the patch component of the latest module version.
'next-build': Uses the latest version and appends +<build> using the short Git commit SHA. (Fails if not in a Git repository)`,
			Required: true,
		},
		includeFlag,
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			version := c.String("version")
			includes := c.StringSlice("include")

			// Read module name from manifest
			manifest, err := modules.ManifestFromFile(moduleManifestFilename)
			if err != nil {
				return err
			}

			// If user specifies --version=next-patch,
			//   we are going to bump the patch automatically from the latest
			if version == "next-patch" {
				version, err = modules.NextPatch(cfg, manifest)
				if err != nil {
					return err
				}
			}
			// If user specifies --version=next-build,
			//   we are going to bump the patch and use the short git commit sha as +build in the semver
			if version == "next-build" {
				version, err = modules.NextPatch(cfg, manifest)
				if err != nil {
					return err
				}
				var commitSha string
				if hash, err := getCurrentCommitSha(); err == nil && len(hash) >= 8 {
					commitSha = hash[0:8]
				} else {
					return fmt.Errorf("Using --version=next-build requires a git repository with a commit. Cannot find commit SHA: %w", err)
				}
				version = fmt.Sprintf("%s+%s", version, commitSha)
			}

			version = strings.TrimPrefix(version, "v")
			if isValid := semver.IsValid(fmt.Sprintf("v%s", version)); !isValid {
				return fmt.Errorf("version %q is not a valid semver", version)
			}

			// Package module files into tar.gz
			tarballFilename, err := modules.Package(manifest, version, includes)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Created module package %q\n", tarballFilename)

			// Open tarball to publish
			tarball, err := os.Open(tarballFilename)
			if err != nil {
				return err
			}
			defer tarball.Close()

			client := api.Client{Config: cfg}
			if err := client.Org(manifest.OrgName).ModuleVersions().Create(manifest.Name, version, tarball); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Published %s/%s@%s\n", manifest.OrgName, manifest.Name, version)
			fmt.Fprintln(os.Stdout, version)
			return nil
		})
	},
}

var ModulesPackage = &cli.Command{
	Name:      "package",
	Usage:     "Package a module",
	UsageText: "nullstone modules package",
	Flags: []cli.Flag{
		includeFlag,
	},
	Action: func(c *cli.Context) error {
		includes := c.StringSlice("include")

		// Read module name from manifest
		manifest, err := modules.ManifestFromFile(moduleManifestFilename)
		if err != nil {
			return err
		}

		tarballFilename, err := modules.Package(manifest, "", includes)
		if err == nil {
			fmt.Printf("created module package %q\n", tarballFilename)
		}
		return err
	},
}
