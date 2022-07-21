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
			Name:     "version",
			Aliases:  []string{"v"},
			Usage:    "Specify a semver version for the module. Specify 'auto' to automatically bump the patch from the latest version.",
			Required: true,
		},
		// TODO: We currently support *.tf, .*tf.tmpl patterns; add support for packaging additional files into the module package
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			version := c.String("version")

			// Read module name from manifest
			manifest, err := modules.ManifestFromFile(moduleManifestFilename)
			if err != nil {
				return err
			}

			// If user specifies --version=auto,
			//   we are going to bump the version automatically from the latest
			if version == "auto" {
				version, err = modules.AutoVersion(cfg, manifest)
				if err != nil {
					return err
				}
			}

			version = strings.TrimPrefix(version, "v")
			if isValid := semver.IsValid(fmt.Sprintf("v%s", version)); !isValid {
				return fmt.Errorf("version %q is not a valid semver", version)
			}

			// Package module files into tar.gz
			tarballFilename, err := modules.Package(manifest, version)
			if err != nil {
				return err
			}
			fmt.Printf("created module package %q\n", tarballFilename)

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
			fmt.Printf("published %s/%s@%s\n", manifest.OrgName, manifest.Name, version)
			return nil
		})
	},
}

var ModulesPackage = &cli.Command{
	Name:      "package",
	Usage:     "Package a module",
	UsageText: "nullstone modules package",
	Flags:     []cli.Flag{
		// TODO: We currently support *.tf, .*tf.tmpl patterns; add support for packaging additional files into the module package
	},
	Action: func(c *cli.Context) error {
		// Read module name from manifest
		manifest, err := modules.ManifestFromFile(moduleManifestFilename)
		if err != nil {
			return err
		}

		tarballFilename, err := modules.Package(manifest, "")
		if err == nil {
			fmt.Printf("created module package %q\n", tarballFilename)
		}
		return err
	},
}
