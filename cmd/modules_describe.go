package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type moduleDescribeOutput struct {
	Module  types.Module         `json:"module"`
	Version *types.ModuleVersion `json:"version,omitempty"`
}

var ModulesDescribe = &cli.Command{
	Name: "describe",
	Description: "Fetches metadata for a module and one of its versions. " +
		"The positional argument accepts `[<org>/]<name>[@<version>]`. " +
		"If the organization is omitted, the current organization is used. " +
		"If a version is specified via `@<version>`, the `--version` flag must not also be set. " +
		"When no version is provided, the latest published version is described.",
	Usage:     "Describe a module and one of its versions",
	UsageText: "nullstone modules describe [<org>/]<name>[@<version>] [--version=<version>] [--format=json|pretty]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Module version to describe. Defaults to the latest published version. Cannot be combined with `@<version>` in the positional argument.",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format. One of: json (default), pretty",
			Value: "json",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()

		if c.NArg() < 1 {
			return fmt.Errorf("module reference is required: [<org>/]<name>[@<version>]")
		}
		if c.NArg() > 1 {
			return fmt.Errorf("only one module reference is allowed")
		}

		ref := c.Args().First()
		orgFromRef, name, versionFromRef, err := parseModuleRef(ref)
		if err != nil {
			return err
		}

		versionFromFlag := c.String("version")
		if versionFromRef != "" && versionFromFlag != "" {
			return fmt.Errorf("cannot specify both `@<version>` in the positional argument and `--version`")
		}
		version := versionFromRef
		if version == "" {
			version = versionFromFlag
		}

		format := strings.ToLower(c.String("format"))
		if format != "json" && format != "pretty" {
			return fmt.Errorf("invalid --format %q: must be json or pretty", format)
		}

		return ProfileAction(c, func(cfg api.Config) error {
			orgName := orgFromRef
			if orgName == "" {
				orgName = cfg.OrgName
			}
			if orgName == "" {
				return ErrMissingOrg
			}

			client := api.Client{Config: cfg}
			module, err := client.Modules().Get(ctx, orgName, name)
			if err != nil {
				return fmt.Errorf("error retrieving module: %w", err)
			}
			if module == nil {
				return fmt.Errorf("module %s/%s does not exist", orgName, name)
			}

			var moduleVersion *types.ModuleVersion
			if version == "" || version == "latest" {
				moduleVersion = module.LatestVersion
				if moduleVersion == nil {
					return fmt.Errorf("module %s/%s has no published versions", orgName, name)
				}
			} else {
				moduleVersion, err = client.ModuleVersions().Get(ctx, orgName, name, version)
				if err != nil {
					return fmt.Errorf("error retrieving module version %s: %w", version, err)
				}
				if moduleVersion == nil {
					return fmt.Errorf("version %s of module %s/%s does not exist", version, orgName, name)
				}
			}

			out := moduleDescribeOutput{Module: *module, Version: moduleVersion}
			if format == "pretty" {
				writeModuleDescribePretty(os.Stdout, out)
				return nil
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		})
	},
}

// parseModuleRef parses a reference of the form `[<org>/]<name>[@<version>]`.
func parseModuleRef(ref string) (org, name, version string, err error) {
	if ref == "" {
		return "", "", "", fmt.Errorf("module reference is empty")
	}
	// Split off the version, if present. Version strings don't contain '@',
	// so a single '@' cleanly separates name from version.
	if idx := strings.Index(ref, "@"); idx >= 0 {
		version = ref[idx+1:]
		ref = ref[:idx]
		if version == "" {
			return "", "", "", fmt.Errorf("missing version after '@'")
		}
	}
	if slashIdx := strings.Index(ref, "/"); slashIdx >= 0 {
		org = ref[:slashIdx]
		name = ref[slashIdx+1:]
	} else {
		name = ref
	}
	if name == "" {
		return "", "", "", fmt.Errorf("module name is required")
	}
	return org, name, version, nil
}

func writeModuleDescribePretty(w *os.File, out moduleDescribeOutput) {
	m := out.Module
	rows := []string{
		fmt.Sprintf("Org|%s", m.OrgName),
		fmt.Sprintf("Name|%s", m.Name),
		fmt.Sprintf("Friendly Name|%s", m.FriendlyName),
	}
	if m.Description != "" {
		rows = append(rows, fmt.Sprintf("Description|%s", m.Description))
	}
	category := string(m.Category)
	if m.Subcategory != "" {
		category = fmt.Sprintf("%s/%s", category, m.Subcategory)
	}
	rows = append(rows, fmt.Sprintf("Category|%s", category))
	if len(m.ProviderTypes) > 0 {
		rows = append(rows, fmt.Sprintf("Providers|%s", strings.Join(m.ProviderTypes, ", ")))
	}
	platform := m.Platform
	if m.Subplatform != "" {
		platform = fmt.Sprintf("%s/%s", platform, m.Subplatform)
	}
	if platform != "" {
		rows = append(rows, fmt.Sprintf("Platform|%s", platform))
	}
	if m.Type != "" {
		rows = append(rows, fmt.Sprintf("Type|%s", m.Type))
	}
	rows = append(rows,
		fmt.Sprintf("Public|%t", m.IsPublic),
		fmt.Sprintf("Status|%s", m.Status),
	)
	if m.SourceUrl != "" {
		rows = append(rows, fmt.Sprintf("Source URL|%s", m.SourceUrl))
	}

	fmt.Fprintln(w, "Module")
	fmt.Fprintln(w, columnize.Format(rows, columnize.DefaultConfig()))

	if out.Version == nil {
		return
	}
	v := out.Version
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Version")
	vrows := []string{
		fmt.Sprintf("Version|%s", v.Version),
		fmt.Sprintf("Tool|%s", v.ToolName),
		fmt.Sprintf("Created At|%s", v.CreatedAt),
	}
	if len(v.Manifest.Providers) > 0 {
		vrows = append(vrows, fmt.Sprintf("Manifest Providers|%s", strings.Join(v.Manifest.Providers, ", ")))
	}
	vrows = append(vrows,
		fmt.Sprintf("Variables|%d", len(v.Manifest.Variables)),
		fmt.Sprintf("Outputs|%d", len(v.Manifest.Outputs)),
		fmt.Sprintf("Connections|%d", len(v.Manifest.Connections)),
		fmt.Sprintf("Env Variables|%d", len(v.Manifest.EnvVariables)),
	)
	fmt.Fprintln(w, columnize.Format(vrows, columnize.DefaultConfig()))
}
