package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
	"os"
	"path"
)

var (
	moduleManifestFilename = path.Join(".nullstone", "module.yml")
	moduleFilePatterns     = []string{
		"*.tf",
		"*.tf.tmpl",
	}
)

var Modules = &cli.Command{
	Name:      "modules",
	Usage:     "View and modify modules",
	UsageText: "nullstone modules [subcommand]",
	Subcommands: []*cli.Command{
		ModulesGenerate,
		ModulesNew,
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
			survey := &moduleSurvey{}
			manifest, err := survey.Ask(cfg)
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

var ModulesNew = &cli.Command{
	Name:      "new",
	Usage:     "Create new module",
	UsageText: "nullstone modules new",
	Flags:     []cli.Flag{},
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
	Usage:     "Publish new version of a module",
	UsageText: "nullstone modules publish --version=<version>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "version",
			Aliases:  []string{"v"},
			Usage:    "Specify a semver version for the module",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			version := c.String("version")
			if isValid := semver.IsValid(version); !isValid {
				return fmt.Errorf("version %q is not a valid semver", version)
			}

			// Read module name from manifest
			manifest, err := modules.ManifestFromFile(moduleManifestFilename)
			if err != nil {
				return err
			}

			// Package module files into tar.gz
			tarballFilename := fmt.Sprintf("%s-%s.tar.gz", manifest.Name, version)
			if err := artifacts.PackageModule(".", tarballFilename, moduleFilePatterns); err != nil {
				return err
			}

			// Open tarball to publish
			tarball, err := os.Open(tarballFilename)
			if err != nil {
				return err
			}
			defer tarball.Close()

			client := api.Client{Config: cfg}
			return client.ModuleVersions().Create(manifest.Name, version, tarball)
		})
	},
}

var ModulesPackage = &cli.Command{
	Name:      "package",
	Usage:     "Package a module",
	UsageText: "nullstone modules package",
	Action: func(c *cli.Context) error {
		// Read module name from manifest
		manifest, err := modules.ManifestFromFile(moduleManifestFilename)
		if err != nil {
			return err
		}

		tarballFilename := fmt.Sprintf("%s.tar.gz", manifest.Name)
		return artifacts.PackageModule(".", tarballFilename, moduleFilePatterns)
	},
}
