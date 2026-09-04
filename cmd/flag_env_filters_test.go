package cmd

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// parseFilters runs the flags through urfave/cli the same way a real invocation does,
// so the tests cover flag declaration and parsing, not just the translation.
func parseFilters(t *testing.T, args ...string) (api.FindEnvironmentsInput, error) {
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

	// The empty value is meaningful ("unset or empty") and must survive to the API as
	// KEY= rather than being dropped as if the filter were absent.
	t.Run("an empty value is kept", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "claim=")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"claim": ""}, filters.Tags)
	})

	t.Run("a value may contain =", func(t *testing.T) {
		filters, err := parseFilters(t, "--tag", "expr=a=b")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"expr": "a=b"}, filters.Tags)
	})

	// Same wording as --env-var so the two flags behave identically.
	t.Run("malformed values error like --env-var", func(t *testing.T) {
		for _, raw := range []string{"claim", "=value", ""} {
			_, err := parseFilters(t, "--tag", raw)
			require.Error(t, err, raw)
			assert.Contains(t, err.Error(), "must be in the form KEY=VALUE")
		}
	})

	t.Run("no tag flag sends no tags", func(t *testing.T) {
		filters, err := parseFilters(t)
		require.NoError(t, err)
		assert.Nil(t, filters.Tags)
	})
}

func TestParseEnvFilters_StatusProdAndName(t *testing.T) {
	t.Run("status defaults to unset and accepts active or archived", func(t *testing.T) {
		none, err := parseFilters(t)
		require.NoError(t, err)
		assert.Empty(t, none.Status)

		archived, err := parseFilters(t, "--status", "Archived")
		require.NoError(t, err)
		assert.Equal(t, types.EnvStatus("archived"), archived.Status)

		_, err = parseFilters(t, "--status", "deleted")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `invalid --status "deleted"`)
	})

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

	t.Run("name is passed through as the search term", func(t *testing.T) {
		filters, err := parseFilters(t, "--name", "pr-*")
		require.NoError(t, err)
		assert.Equal(t, "pr-*", filters.Search)
	})
}

// No flags must produce the zero input, which the API treats as today's unfiltered list.
func TestParseEnvFilters_NoFlagsIsZero(t *testing.T) {
	filters, err := parseFilters(t)
	require.NoError(t, err)
	assert.Equal(t, api.FindEnvironmentsInput{}, filters)
}
