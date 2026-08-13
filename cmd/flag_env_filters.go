package cmd

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// EnvTypeFlag filters by environment type using the friendly names the --detail table
// prints, not the raw PipelineEnv/PreviewEnv enum.
var EnvTypeFlag = &cli.StringSliceFlag{
	Name: "type",
	Usage: `Only show environments of this type: pipeline, preview, previews-shared, or global.
		Can be specified multiple times to show more than one type.`,
}

// EnvTagFlag filters by the open user tag keyspace on an environment.
var EnvTagFlag = &cli.StringSliceFlag{
	Name: "tag",
	Usage: `Only show environments whose tags match KEY=VALUE.
		Can be specified multiple times; an environment must match every tag given.
		An empty value (--tag claim=) matches environments where the tag is unset, absent, or empty,
		which is how you find environments that haven't been tagged yet.`,
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
	Usage: `Only show environments matching this name. A pattern containing *, ? or [ is matched
		as a glob against the whole name (--name='pr-*'); anything else matches as a substring.`,
}

// EnvFilterFlags is the full set, for a command that wants all of them.
var EnvFilterFlags = []cli.Flag{
	EnvTypeFlag,
	EnvTagFlag,
	EnvProdFlag,
	EnvNonProdFlag,
	EnvNameFlag,
}

var envTypesByFriendlyName = map[string]types.EnvironmentType{
	"pipeline":        types.EnvTypePipeline,
	"preview":         types.EnvTypePreview,
	"previews-shared": types.EnvTypePreviewsShared,
	"global":          types.EnvTypeGlobal,
}

// EnvFilters filters an environment list client-side. The list endpoint takes no query
// params, and lists at this scale don't justify server-side filtering. A zero EnvFilters
// matches everything.
//
// There is no status filter: the list endpoint returns active environments only, so
// filtering on status client-side could never surface an archived environment.
type EnvFilters struct {
	// Types matches an environment with any one of these types (OR).
	Types []types.EnvironmentType
	// Tags matches an environment carrying every one of these tags (AND).
	Tags map[string]string
	// IsProd matches on the prod flag. nil matches either.
	IsProd *bool
	// Name matches the environment name as a glob or a substring.
	Name string
}

// ParseEnvFilters reads the filter flags off the context, rejecting unusable input
// before any environments are fetched.
func ParseEnvFilters(c *cli.Context) (EnvFilters, error) {
	var filters EnvFilters

	for _, raw := range c.StringSlice(EnvTypeFlag.Name) {
		envType, ok := envTypesByFriendlyName[strings.ToLower(strings.TrimSpace(raw))]
		if !ok {
			return filters, fmt.Errorf("invalid --type %q: must be one of %s", raw, sortedKeys(envTypesByFriendlyName))
		}
		filters.Types = append(filters.Types, envType)
	}

	if raw := c.StringSlice(EnvTagFlag.Name); len(raw) > 0 {
		filters.Tags = make(map[string]string, len(raw))
		for _, kvp := range raw {
			tokens := strings.SplitN(kvp, "=", 2)
			if len(tokens) < 2 || tokens[0] == "" {
				return filters, fmt.Errorf("invalid --tag %q: must be in the form KEY=VALUE", kvp)
			}
			filters.Tags[tokens[0]] = tokens[1]
		}
	}

	isProd, isNonProd := c.Bool(EnvProdFlag.Name), c.Bool(EnvNonProdFlag.Name)
	if isProd && isNonProd {
		return filters, fmt.Errorf("--prod and --non-prod cannot be used together")
	}
	if isProd || isNonProd {
		filters.IsProd = &isProd
	}

	filters.Name = c.String(EnvNameFlag.Name)
	if isGlob(filters.Name) {
		if _, err := path.Match(filters.Name, ""); err != nil {
			return filters, fmt.Errorf("invalid --name %q: %w", filters.Name, err)
		}
	}

	return filters, nil
}

// Apply returns the environments that match every filter, preserving order.
func (f EnvFilters) Apply(envs []*types.Environment) []*types.Environment {
	result := make([]*types.Environment, 0, len(envs))
	for _, env := range envs {
		if env != nil && f.Matches(*env) {
			result = append(result, env)
		}
	}
	return result
}

// Matches reports whether an environment satisfies every filter. Filters AND together;
// repeated values within --type OR together.
func (f EnvFilters) Matches(env types.Environment) bool {
	if len(f.Types) > 0 && !containsEnvType(f.Types, env.Type) {
		return false
	}
	if f.IsProd != nil && env.IsProd != *f.IsProd {
		return false
	}
	if f.Name != "" && !matchesEnvName(f.Name, env.Name) {
		return false
	}
	for key, value := range f.Tags {
		if !matchesEnvTag(env.Tags, key, value) {
			return false
		}
	}
	return true
}

// matchesEnvTag reports whether an environment's tags satisfy one KEY=VALUE filter.
//
// A non-empty value requires the key to be present and equal. An empty value means
// "unset or empty" rather than "literally the empty string", so it matches when the
// environment has no tags at all, when the key is absent, and when the key is present
// with an empty value. Those three are what make the filter useful for finding
// environments nobody has tagged yet.
func matchesEnvTag(tags map[string]string, key, value string) bool {
	existing, ok := tags[key]
	if value == "" {
		return !ok || existing == ""
	}
	return ok && existing == value
}

func matchesEnvName(pattern, name string) bool {
	if isGlob(pattern) {
		matched, err := path.Match(strings.ToLower(pattern), strings.ToLower(name))
		return err == nil && matched
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(pattern))
}

func isGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func containsEnvType(envTypes []types.EnvironmentType, envType types.EnvironmentType) bool {
	for _, cur := range envTypes {
		if cur == envType {
			return true
		}
	}
	return false
}

func sortedKeys[T any](m map[string]T) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
