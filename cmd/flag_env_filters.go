package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// EnvTypeFlag filters by environment type using the friendly names the --detail table
// prints, not the raw PipelineEnv/PreviewEnv enum.
var EnvTypeFlag = &cli.StringSliceFlag{
	Name: "type",
	Usage: `Only show environments of this type: pipeline, preview, previews-shared, or global.
		Can be specified multiple times to show more than one type.`,
}

// EnvTagFilterFlag filters by the open user tag keyspace on an environment. Note its
// empty-value semantics differ from the write-side --tag in flag_env_tags.go: here an
// empty value means "unset or empty", there it sets the tag to the empty string.
var EnvTagFilterFlag = &cli.StringSliceFlag{
	Name: "tag",
	Usage: `Only show environments whose tags match KEY=VALUE.
		Can be specified multiple times; an environment must match every tag given.
		An empty value (--tag claim=) matches environments where the tag is unset, absent, or empty,
		which is how you find environments that haven't been tagged yet.
		A bare key (--tag claim) matches environments that have the tag with any value.`,
}

var EnvStatusFlag = &cli.StringFlag{
	Name:  "status",
	Usage: "Only show environments with this status: active (the default) or archived.",
}

var EnvProdFlag = &cli.BoolFlag{
	Name:  "prod",
	Usage: "Only show production environments. Cannot be combined with --non-prod.",
}

var EnvNonProdFlag = &cli.BoolFlag{
	Name:  "non-prod",
	Usage: "Only show non-production environments. Cannot be combined with --prod.",
}

var EnvNameFlag = &cli.StringFlag{
	Name: "name",
	Usage: `Only show environments matching this name, case-insensitively. A pattern containing * or ?
		is matched against the whole name (--name='pr-*'); anything else matches as a substring.`,
}

// EnvFilterFlags is the full set, for a command that wants all of them.
var EnvFilterFlags = []cli.Flag{
	EnvTypeFlag,
	EnvTagFilterFlag,
	EnvStatusFlag,
	EnvProdFlag,
	EnvNonProdFlag,
	EnvNameFlag,
}

var envStatusesByName = map[string]types.EnvStatus{
	"active":   types.EnvStatusActive,
	"archived": types.EnvStatusArchived,
}

// ParseEnvFilters reads the filter flags off the context into the query the API takes,
// rejecting unusable input before any request is made. Filtering happens server-side, so
// the flags only translate; the match semantics live in the API.
func ParseEnvFilters(c *cli.Context) (api.FindEnvironmentsInput, error) {
	var filters api.FindEnvironmentsInput

	for _, raw := range c.StringSlice(EnvTypeFlag.Name) {
		envType, err := parseEnvType(raw)
		if err != nil {
			return filters, err
		}
		filters.Types = append(filters.Types, envType)
	}

	if raw := c.StringSlice(EnvTagFilterFlag.Name); len(raw) > 0 {
		for _, kvp := range raw {
			tokens := strings.SplitN(kvp, "=", 2)
			if tokens[0] == "" {
				return filters, fmt.Errorf("invalid --tag %q: must be in the form KEY=VALUE or KEY", kvp)
			}
			if len(tokens) < 2 {
				// bare key: the tag must be present, any value
				filters.HasTags = append(filters.HasTags, tokens[0])
				continue
			}
			if filters.Tags == nil {
				filters.Tags = make(map[string]string, len(raw))
			}
			filters.Tags[tokens[0]] = tokens[1]
		}
	}

	if raw := c.String(EnvStatusFlag.Name); raw != "" {
		status, ok := envStatusesByName[strings.ToLower(strings.TrimSpace(raw))]
		if !ok {
			return filters, fmt.Errorf("invalid --status %q: must be one of active, archived", raw)
		}
		filters.Status = status
	}

	isProd, isNonProd := c.Bool(EnvProdFlag.Name), c.Bool(EnvNonProdFlag.Name)
	if isProd && isNonProd {
		return filters, fmt.Errorf("--prod and --non-prod cannot be used together")
	}
	if isProd || isNonProd {
		filters.IsProd = &isProd
	}

	filters.Search = c.String(EnvNameFlag.Name)
	return filters, nil
}
