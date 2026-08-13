package cmd

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

// parseUpdate runs args through urfave/cli exactly as EnvsUpdate declares them, so the
// tests cover flag declaration and IsSet behavior, not just the builder.
func parseUpdate(t *testing.T, args ...string) (*api.UpdateEnvironmentInput, map[string]*string, error) {
	t.Helper()

	set := flag.NewFlagSet("update", flag.ContinueOnError)
	for _, f := range EnvsUpdate.Flags {
		require.NoError(t, f.Apply(set))
	}
	require.NoError(t, set.Parse(args))
	return parseEnvUpdate(cli.NewContext(cli.NewApp(), set, nil))
}

func TestParseEnvUpdate_Tags(t *testing.T) {
	t.Run("sets a tag", func(t *testing.T) {
		input, patch, err := parseUpdate(t, "--tag", "claim=brad")
		require.NoError(t, err)
		assert.Nil(t, input, "no attribute update when only tags change")
		require.Contains(t, patch, "claim")
		require.NotNil(t, patch["claim"])
		assert.Equal(t, "brad", *patch["claim"])
	})

	t.Run("removes a tag", func(t *testing.T) {
		_, patch, err := parseUpdate(t, "--remove-tag", "claim")
		require.NoError(t, err)
		require.Contains(t, patch, "claim")
		assert.Nil(t, patch["claim"], "a nil value clears the key")
	})

	// The distinction NUL-174 preserves: "" keeps the key, removal drops it.
	t.Run("an empty value is not a removal", func(t *testing.T) {
		_, patch, err := parseUpdate(t, "--tag", "claim=")
		require.NoError(t, err)
		require.NotNil(t, patch["claim"])
		assert.Equal(t, "", *patch["claim"])
	})

	t.Run("sets and removes in one call", func(t *testing.T) {
		_, patch, err := parseUpdate(t, "--tag", "tier=gold", "--remove-tag", "claim")
		require.NoError(t, err)
		require.Len(t, patch, 2)
		assert.Equal(t, "gold", *patch["tier"])
		assert.Nil(t, patch["claim"])
	})

	t.Run("repeated tags accumulate distinctly", func(t *testing.T) {
		_, patch, err := parseUpdate(t, "--tag", "a=1", "--tag", "b=2")
		require.NoError(t, err)
		assert.Equal(t, "1", *patch["a"])
		assert.Equal(t, "2", *patch["b"], "each key keeps its own value, not the last one")
	})

	t.Run("setting and removing the same key errors", func(t *testing.T) {
		_, _, err := parseUpdate(t, "--tag", "claim=brad", "--remove-tag", "claim")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both set and removed")
	})

	t.Run("a malformed tag errors like --env-var", func(t *testing.T) {
		_, _, err := parseUpdate(t, "--tag", "claim")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be in the form KEY=VALUE")
	})

	t.Run("a remove-tag with a value errors", func(t *testing.T) {
		_, _, err := parseUpdate(t, "--remove-tag", "claim=brad")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "give the key only")
	})
}

func TestParseEnvUpdate_Attributes(t *testing.T) {
	t.Run("renames and sanitizes", func(t *testing.T) {
		input, patch, err := parseUpdate(t, "--name", "My Env")
		require.NoError(t, err)
		assert.Nil(t, patch)
		require.NotNil(t, input.Name)
		assert.Equal(t, "my-env", *input.Name)
	})

	t.Run("sets a description", func(t *testing.T) {
		input, _, err := parseUpdate(t, "--description", "payments sandbox")
		require.NoError(t, err)
		require.NotNil(t, input.Metadata)
		require.NotNil(t, input.Metadata.Description)
		assert.Equal(t, "payments sandbox", *input.Metadata.Description)
	})

	// An explicitly empty --description clears it; omitting the flag leaves it alone.
	t.Run("an empty description clears it", func(t *testing.T) {
		input, _, err := parseUpdate(t, "--description", "", "--tag", "a=1")
		require.NoError(t, err)
		require.NotNil(t, input)
		require.NotNil(t, input.Metadata)
		assert.Equal(t, "", *input.Metadata.Description)
	})

	t.Run("an omitted description is left untouched", func(t *testing.T) {
		input, _, err := parseUpdate(t, "--name", "dev")
		require.NoError(t, err)
		assert.Nil(t, input.Metadata, "nil metadata means the API leaves it alone")
	})

	t.Run("prod flags", func(t *testing.T) {
		input, _, err := parseUpdate(t, "--prod")
		require.NoError(t, err)
		require.NotNil(t, input.IsProd)
		assert.True(t, *input.IsProd)

		input, _, err = parseUpdate(t, "--non-prod")
		require.NoError(t, err)
		require.NotNil(t, input.IsProd)
		assert.False(t, *input.IsProd)
	})

	t.Run("prod and non-prod are mutually exclusive", func(t *testing.T) {
		_, _, err := parseUpdate(t, "--prod", "--non-prod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be used together")
	})
}

func TestParseEnvUpdate_NothingToUpdate(t *testing.T) {
	_, _, err := parseUpdate(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to update")
}
