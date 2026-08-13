package cmd

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// parseFilters runs the flags through urfave/cli the same way a real invocation does,
// so the tests cover flag declaration and parsing, not just the matcher.
func parseFilters(t *testing.T, args ...string) (EnvFilters, error) {
	t.Helper()

	set := flag.NewFlagSet("list", flag.ContinueOnError)
	for _, f := range EnvFilterFlags {
		require.NoError(t, f.Apply(set))
	}
	require.NoError(t, set.Parse(args))
	return ParseEnvFilters(cli.NewContext(cli.NewApp(), set, nil))
}

func TestParseEnvFilters_Type(t *testing.T) {
	t.Run("friendly names map to the enum", func(t *testing.T) {
		filters, err := parseFilters(t, "--type", "preview")
		require.NoError(t, err)
		assert.Equal(t, []types.EnvironmentType{types.EnvTypePreview}, filters.Types)
	})

	t.Run("is case-insensitive", func(t *testing.T) {
		filters, err := parseFilters(t, "--type", "PreViews-Shared")
		require.NoError(t, err)
		assert.Equal(t, []types.EnvironmentType{types.EnvTypePreviewsShared}, filters.Types)
	})

	t.Run("repeats accumulate", func(t *testing.T) {
		filters, err := parseFilters(t, "--type", "preview", "--type", "global")
		require.NoError(t, err)
		assert.Equal(t, []types.EnvironmentType{types.EnvTypePreview, types.EnvTypeGlobal}, filters.Types)
	})

	// An unknown value must error rather than silently return zero rows.
	t.Run("an unknown type errors and lists the valid set", func(t *testing.T) {
		_, err := parseFilters(t, "--type", "staging")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `invalid --type "staging"`)
		assert.Contains(t, err.Error(), "global, pipeline, preview, previews-shared")
	})
}

func TestParseEnvFilters_Tag(t *testing.T) {
	t.Run("parses KEY=VALUE", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "claim=brad")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"claim": "brad"}, filters.Tags)
	})

	t.Run("repeats accumulate", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "claim=brad", "--tag", "tier=gold")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"claim": "brad", "tier": "gold"}, filters.Tags)
	})

	t.Run("keeps an empty value", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "claim=")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"claim": ""}, filters.Tags)
	})

	t.Run("a value containing = is kept intact", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "expr=a=b")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"expr": "a=b"}, filters.Tags)
	})

	// Same wording as --env-var, so the two flags behave identically.
	t.Run("errors on a malformed value", func(t *testing.T) {
		for _, raw := range []string{"claim", "=brad"} {
			_, err := parseFilters(t, "--tag", raw)
			require.Error(t, err, raw)
			assert.Contains(t, err.Error(), "must be in the form KEY=VALUE")
		}
	})
}

func TestParseEnvFilters_ProdAndName(t *testing.T) {
	t.Run("prod and non-prod are mutually exclusive", func(t *testing.T) {
		_, err := parseFilters(t, "--prod", "--non-prod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be used together")
	})

	t.Run("prod flags set the tri-state", func(t *testing.T) {
		none, err := parseFilters(t)
		require.NoError(t, err)
		assert.Nil(t, none.IsProd)

		prod, err := parseFilters(t, "--prod")
		require.NoError(t, err)
		require.NotNil(t, prod.IsProd)
		assert.True(t, *prod.IsProd)

		nonProd, err := parseFilters(t, "--non-prod")
		require.NoError(t, err)
		require.NotNil(t, nonProd.IsProd)
		assert.False(t, *nonProd.IsProd)
	})

	t.Run("an invalid glob errors", func(t *testing.T) {
		_, err := parseFilters(t, "--name", "pr-[")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid --name")
	})
}

func TestEnvFilters_MatchesTag(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		key  string
		want bool
	}{
		{name: "value matches", tags: map[string]string{"claim": "brad"}, key: "claim=brad", want: true},
		{name: "value differs", tags: map[string]string{"claim": "alex"}, key: "claim=brad", want: false},
		{name: "key absent", tags: map[string]string{"tier": "gold"}, key: "claim=brad", want: false},
		{name: "no tags at all", tags: nil, key: "claim=brad", want: false},

		// The three branches of the empty-value case: "unset or empty", not "literally empty".
		{name: "empty matches nil tags", tags: nil, key: "claim=", want: true},
		{name: "empty matches an empty tag map", tags: map[string]string{}, key: "claim=", want: true},
		{name: "empty matches an absent key", tags: map[string]string{"tier": "gold"}, key: "claim=", want: true},
		{name: "empty matches a key present as empty", tags: map[string]string{"claim": ""}, key: "claim=", want: true},
		{name: "empty does not match a set key", tags: map[string]string{"claim": "brad"}, key: "claim=", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filters, err := parseFilters(t, "--tag", test.key)
			require.NoError(t, err)

			got := filters.Matches(types.Environment{Status: types.EnvStatusActive, Tags: test.tags})

			assert.Equal(t, test.want, got)
		})
	}
}

func TestEnvFilters_Apply(t *testing.T) {
	prod := &types.Environment{Name: "prod", Type: types.EnvTypePipeline, Status: types.EnvStatusActive, IsProd: true, Tags: map[string]string{"tier": "gold"}}
	dev := &types.Environment{Name: "dev", Type: types.EnvTypePipeline, Status: types.EnvStatusActive}
	pr1 := &types.Environment{Name: "pr-1", Type: types.EnvTypePreview, Status: types.EnvStatusActive, Tags: map[string]string{"tier": "gold"}}
	pr2 := &types.Environment{Name: "pr-2", Type: types.EnvTypePreview, Status: types.EnvStatusActive}
	global := &types.Environment{Name: "global", Type: types.EnvTypeGlobal, Status: types.EnvStatusActive}
	all := []*types.Environment{prod, dev, pr1, pr2, global}

	tests := []struct {
		name string
		args []string
		want []*types.Environment
	}{
		{name: "no filters returns everything in order", args: nil, want: all},
		{name: "single type", args: []string{"--type", "preview"}, want: []*types.Environment{pr1, pr2}},
		{name: "repeated type ORs", args: []string{"--type", "preview", "--type", "global"}, want: []*types.Environment{pr1, pr2, global}},
		{name: "prod", args: []string{"--prod"}, want: []*types.Environment{prod}},
		{name: "non-prod", args: []string{"--non-prod"}, want: []*types.Environment{dev, pr1, pr2, global}},
		{name: "name glob", args: []string{"--name", "pr-*"}, want: []*types.Environment{pr1, pr2}},
		{name: "name substring", args: []string{"--name", "ro"}, want: []*types.Environment{prod}},
		{name: "flags AND together", args: []string{"--type", "preview", "--tag", "tier=gold"}, want: []*types.Environment{pr1}},
		{name: "repeated tags AND together", args: []string{"--tag", "tier=gold", "--tag", "claim="}, want: []*types.Environment{prod, pr1}},
		{name: "no matches returns empty", args: []string{"--type", "preview", "--prod"}, want: []*types.Environment{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filters, err := parseFilters(t, test.args...)
			require.NoError(t, err)

			got := filters.Apply(all)

			assert.Equal(t, test.want, got)
		})
	}
}

// No filters must leave today's output byte-for-byte identical.
func TestEnvFilters_NoFiltersIsANoOp(t *testing.T) {
	envs := []*types.Environment{
		{Name: "dev", Type: types.EnvTypePipeline, Status: types.EnvStatusActive},
		{Name: "prod", Type: types.EnvTypePipeline, Status: types.EnvStatusActive},
	}

	filters, err := parseFilters(t)
	require.NoError(t, err)

	assert.Equal(t, envs, filters.Apply(envs))
}
