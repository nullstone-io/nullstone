package cmd

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// parseEnvVars runs args through urfave/cli the same way a real invocation does,
// so the tests cover the StringSliceFlag parsing path (where comma-splitting
// happens), not just ParseEnvVars.
func parseEnvVars(t *testing.T, args ...string) (map[string]string, error) {
	t.Helper()

	set := flag.NewFlagSet("deploy", flag.ContinueOnError)
	require.NoError(t, EnvVarFlag.Apply(set))
	require.NoError(t, set.Parse(args))
	return ParseEnvVars(cli.NewContext(cli.NewApp(), set, nil))
}

func TestParseEnvVars(t *testing.T) {
	t.Run("parses KEY=VALUE", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", "FOO=bar")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"FOO": "bar"}, envVars)
	})

	t.Run("repeats accumulate", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", "FOO=bar", "--env-var", "BAZ=qux")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"FOO": "bar", "BAZ": "qux"}, envVars)
	})

	t.Run("keeps an empty value", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", "FOO=")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"FOO": ""}, envVars)
	})

	t.Run("a value containing = is kept intact", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", "EXPR=a=b")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"EXPR": "a=b"}, envVars)
	})

	t.Run("errors on a malformed value", func(t *testing.T) {
		for _, raw := range []string{"FOO", "=bar"} {
			_, err := parseEnvVars(t, "--env-var", raw)
			require.Error(t, err, raw)
			assert.Contains(t, err.Error(), "must be in the form KEY=VALUE")
		}
	})
}

// Issue #673: StringSliceFlag splits values on commas before ParseEnvVars runs,
// corrupting any value that legitimately contains a comma (JSON, CSV lists,
// connection strings). These tests assert the correct behavior and fail until
// the comma-splitting is disabled for --env-var.
func TestParseEnvVars_ValuesWithCommas(t *testing.T) {
	t.Run("a plain value containing commas is kept intact", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", "HOSTS=a.example.com,b.example.com")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"HOSTS": "a.example.com,b.example.com"}, envVars)
	})

	// Failure mode 1 from the issue: the fragment after the comma has no "=",
	// so today this errors with a misleading "must be in the form KEY=VALUE".
	t.Run("a JSON value with multiple keys is kept intact", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", `JSON={"a":"b","c":"d"}`)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"JSON": `{"a":"b","c":"d"}`}, envVars)
	})

	// Failure mode 2 from the issue: the fragment after the comma happens to
	// contain "=", so today it parses "successfully" — truncating JSON to
	// {"a":"b" and inventing a spurious variable named `"c`.
	t.Run("a JSON value whose fragment contains = is not silently corrupted", func(t *testing.T) {
		envVars, err := parseEnvVars(t, "--env-var", `JSON={"a":"b","c=1":"d"}`)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"JSON": `{"a":"b","c=1":"d"}`}, envVars)
	})

	// Comma-containing values must not bleed into neighboring repeats.
	t.Run("repeats with comma values stay separate and intact", func(t *testing.T) {
		envVars, err := parseEnvVars(t,
			"--env-var", `JSON={"a":"b","c":"d"}`,
			"--env-var", "FOO=bar")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"JSON": `{"a":"b","c":"d"}`,
			"FOO":  "bar",
		}, envVars)
	})
}
