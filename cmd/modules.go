package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
)

var (
	moduleManifestFilename = path.Join(".nullstone", "module.yml")
	includeFlag            = &cli.StringSliceFlag{
		Name: "include",
		Usage: `Specify additional file patterns to package. 
By default, this command includes *.tf, *.tf.tmpl, and 'README.md'. 
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
		ModulesList,
		ModulesRegister,
		ModulesPublish,
		ModulesPackage,
	},
}

var ModulesList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of modules in the Nullstone registry for the current organization.",
	Usage:       "List modules",
	UsageText:   "nullstone modules list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "detailed",
			Usage: "Use this flag to show detailed information for each module",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Filter modules whose name contains this value",
		},
		&cli.StringFlag{
			Name:  "category",
			Usage: "Filter modules by category. Known values: app, capability, datastore, ingress, subdomain, domain, cluster, cluster-namespace, network, block",
		},
		&cli.StringFlag{
			Name:  "subcategory",
			Usage: "Filter modules by subcategory. Known values — app: container, serverless, static-site, server; capability: ingress, datastores, secrets, sidecars, events, telemetry",
		},
		&cli.StringFlag{
			Name:  "provider",
			Usage: "Filter modules by provider type. Known values: aws, gcp, azure",
		},
		&cli.StringFlag{
			Name:  "platform",
			Usage: "Filter modules by platform",
		},
		&cli.StringFlag{
			Name:  "subplatform",
			Usage: "Filter modules by subplatform",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			allModules, err := client.Modules().List(ctx, cfg.OrgName)
			if err != nil {
				return fmt.Errorf("error listing modules: %w", err)
			}

			nameFilter := c.String("name")
			categoryFilter := c.String("category")
			subcategoryFilter := c.String("subcategory")
			providerFilter := c.String("provider")
			platformFilter := c.String("platform")
			subplatformFilter := c.String("subplatform")

			filtered := allModules[:0]
			for _, module := range allModules {
				if nameFilter != "" && !strings.Contains(module.Name, nameFilter) {
					continue
				}
				if categoryFilter != "" && string(module.Category) != categoryFilter {
					continue
				}
				if subcategoryFilter != "" && string(module.Subcategory) != subcategoryFilter {
					continue
				}
				if providerFilter != "" && !slices.Contains(module.ProviderTypes, providerFilter) {
					continue
				}
				if platformFilter != "" && module.Platform != platformFilter {
					continue
				}
				if subplatformFilter != "" && module.Subplatform != subplatformFilter {
					continue
				}
				filtered = append(filtered, module)
			}

			if c.IsSet("detailed") {
				moduleDetails := make([]string, len(filtered)+1)
				moduleDetails[0] = "org|name|friendly-name|repo|category|provider|platform|latest-version"
				for i, module := range filtered {
					category := string(module.Category)
					if module.Subcategory != "" {
						category = fmt.Sprintf("%s/%s", category, module.Subcategory)
					}
					platform := module.Platform
					if module.Subplatform != "" {
						platform = fmt.Sprintf("%s/%s", platform, module.Subplatform)
					}
					latestVersion := "<no-versions>"
					if module.LatestVersion != nil {
						latestVersion = module.LatestVersion.Version
					}
					moduleDetails[i+1] = fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
						module.OrgName,
						module.Name,
						module.FriendlyName,
						module.SourceUrl,
						category,
						strings.Join(module.ProviderTypes, ","),
						platform,
						latestVersion,
					)
				}
				fmt.Println(columnize.Format(moduleDetails, columnize.DefaultConfig()))
			} else {
				for _, module := range filtered {
					fmt.Printf("%s/%s\n", module.OrgName, module.Name)
				}
			}

			return nil
		})
	},
}

var ModulesGenerate = &cli.Command{
	Name: "generate",
	Description: "Generates a nullstone manifest file for your module in the current directory. " +
		"You will be asked a series of questions in order to collect the information needed to describe a Nullstone module. " +
		"Optionally, you can also register the module in the Nullstone registry by passing the `--register` flag.",
	Usage:     "Generate new module manifest (and optionally register)",
	UsageText: "nullstone modules generate [--register] [--manifest-only]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "register",
			Usage: "Register the module in the Nullstone registry after generating the manifest file.",
		},
		&cli.BoolFlag{
			Name:  "manifest-only",
			Usage: "Only generate the module manifest file, do not register or generate Terraform.",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			existing, _ := modules.ManifestFromFile(moduleManifestFilename)
			survey := &moduleSurvey{}
			manifest, err := survey.Ask(cfg, existing)
			if err != nil {
				return err
			}
			if err := modules.WriteManifestToFile(*manifest, moduleManifestFilename); err != nil {
				return err
			}
			fmt.Printf("generated module manifest file to %s\n", moduleManifestFilename)

			if c.IsSet("manifest-only") {
				return nil
			}

			if err := modules.Generate(manifest); err != nil {
				return err
			}
			fmt.Printf("generated base Terraform\n")

			if c.IsSet("register") {
				module, err := modules.Register(ctx, cfg, manifest)
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
	Name:        "register",
	Description: "Registers a module in the Nullstone registry. The information in .nullstone/module.yml will be used as the details for the new module.",
	Usage:       "Register module from .nullstone/module.yml",
	UsageText:   "nullstone modules register",
	Flags:       []cli.Flag{},
	Aliases:     []string{"new"},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			manifest, err := modules.ManifestFromFile(moduleManifestFilename)
			if err != nil {
				return err
			}

			module, err := modules.Register(ctx, cfg, manifest)
			if err != nil {
				return err
			}
			fmt.Printf("registered %s/%s\n", module.OrgName, module.Name)
			return nil
		})
	},
}

var ModulesPublish = &cli.Command{
	Name:        "publish",
	Description: "Publishes a new version for a module in the Nullstone registry. Provide a specific semver version using the `--version` parameter.",
	Usage:       "Package and publish new version of a module",
	UsageText:   "nullstone modules publish --version=<version>",
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
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			version := c.String("version")
			includes := c.StringSlice("include")

			output, err := modules.Publish(ctx, cfg, modules.PublishInput{
				Version:  version,
				Includes: includes,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, output.Version)
			return nil
		})
	},
}

var ModulesPackage = &cli.Command{
	Name:        "package",
	Description: "Package all the module contents for a Nullstone module into a tarball but do not publish to the registry.",
	Usage:       "Package a module",
	UsageText:   "nullstone modules package",
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

		tarballFilename, err := modules.Package(manifest, "", append(includes, manifest.Includes...))
		if err == nil {
			fmt.Printf("created module package %q\n", tarballFilename)
		}
		return err
	},
}
