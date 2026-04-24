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

var ModulesFind = &cli.Command{
	Name: "find",
	Description: "Search the entire Nullstone module registry across Nullstone-official, your org, and community modules. " +
		"To list modules owned by your organization, use `nullstone modules list`. " +
		"When `--contributor` is not specified, results include Nullstone-official and your organization's modules.",
	Usage:     "Search the Nullstone module registry",
	UsageText: "nullstone modules find [--category=<category>] [--subcategory=<subcategory>] [--provider=<provider>] [--platform=<platform>] [--subplatform=<subplatform>] [--name=<name>] [--contributor=<contributor>]... [--format=table|json]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "category",
			Usage: "Filter modules by category. Known values: app, capability, datastore, ingress, subdomain, domain, cluster, cluster-namespace, network, block",
		},
		&cli.StringFlag{
			Name:  "subcategory",
			Usage: "Filter modules by subcategory. Requires --category. Known values — app: container, serverless, static-site, server; capability: ingress, datastores, secrets, sidecars, events, telemetry",
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
			Usage: "Filter modules by subplatform. Requires --platform.",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Fuzzy match modules by name",
		},
		&cli.StringSliceFlag{
			Name:  "contributor",
			Usage: "Filter by contributor. Repeat the flag to include multiple. Allowed values: nullstone-official, my-org, community. Defaults to nullstone-official,my-org.",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format. One of: table (default), json",
			Value: "table",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()

		if c.IsSet("subcategory") && !c.IsSet("category") {
			return fmt.Errorf("--subcategory requires --category")
		}
		if c.IsSet("subplatform") && !c.IsSet("platform") {
			return fmt.Errorf("--subplatform requires --platform")
		}

		format := strings.ToLower(c.String("format"))
		if format != "table" && format != "json" {
			return fmt.Errorf("invalid --format %q: must be table or json", format)
		}

		contributors, err := parseContributors(c.StringSlice("contributor"))
		if err != nil {
			return err
		}

		input := api.FindModulesInput{Contributor: contributors}
		if c.IsSet("category") {
			v := c.String("category")
			input.Category = &v
		}
		if c.IsSet("subcategory") {
			v := c.String("subcategory")
			input.Subcategory = &v
		}
		if c.IsSet("provider") {
			v := c.String("provider")
			input.Provider = &v
		}
		if c.IsSet("platform") {
			v := c.String("platform")
			input.Platform = &v
		}
		if c.IsSet("subplatform") {
			v := c.String("subplatform")
			input.Subplatform = &v
		}
		if c.IsSet("name") {
			v := c.String("name")
			input.Name = &v
		}

		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			found, err := client.Modules().Find(ctx, cfg.OrgName, input)
			if err != nil {
				return fmt.Errorf("error searching modules: %w", err)
			}

			if format == "json" {
				if found == nil {
					found = []types.Module{}
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(found)
			}
			writeModulesFindTable(found)
			return nil
		})
	},
}

func parseContributors(raw []string) ([]types.Contributor, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	valid := map[string]types.Contributor{}
	for _, c := range types.AllContributors {
		valid[string(c)] = c
	}
	out := make([]types.Contributor, 0, len(raw))
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		v, ok := valid[r]
		if !ok {
			allowed := make([]string, 0, len(types.AllContributors))
			for _, c := range types.AllContributors {
				allowed = append(allowed, string(c))
			}
			return nil, fmt.Errorf("invalid --contributor %q: must be one of %s", r, strings.Join(allowed, ", "))
		}
		out = append(out, v)
	}
	return out, nil
}

func writeModulesFindTable(modules []types.Module) {
	rows := make([]string, 0, len(modules)+1)
	rows = append(rows, "org|name|category|provider|platform|latest-version")
	for _, m := range modules {
		category := string(m.Category)
		if m.Subcategory != "" {
			category = fmt.Sprintf("%s/%s", category, m.Subcategory)
		}
		platform := m.Platform
		if m.Subplatform != "" {
			platform = fmt.Sprintf("%s/%s", platform, m.Subplatform)
		}
		latest := "<no-versions>"
		if m.LatestVersion != nil {
			latest = m.LatestVersion.Version
		}
		rows = append(rows, fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			m.OrgName,
			m.Name,
			category,
			strings.Join(m.ProviderTypes, ","),
			platform,
			latest,
		))
	}
	fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
}
