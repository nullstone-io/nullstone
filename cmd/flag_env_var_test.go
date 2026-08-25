package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// parseEnvVars runs args through urfave/cli the same way a real invocation does
// The tests cover the StringSliceFlag parsing path (where comma-splitting would happen), not just ParseEnvVars.
// DisableSliceFlagSeparator mirrors app.Build (it is app-global config, applied during App.Setup).
func parseEnvVars(t *testing.T, args ...string) (map[string]string, error) {
	t.Helper()

	var envVars map[string]string
	var parseErr error
	cliApp := &cli.App{
		DisableSliceFlagSeparator: true,
		Flags:                     []cli.Flag{EnvVarFlag},
		Action: func(c *cli.Context) error {
			envVars, parseErr = ParseEnvVars(c)
			return nil
		},
	}
	require.NoError(t, cliApp.Run(append([]string{"deploy"}, args...)))
	return envVars, parseErr
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

// Issue #673: StringSliceFlag used to split values on commas before ParseEnvVars
// ran, corrupting any value that legitimately contains a comma (JSON, CSV lists,
// connection strings). DisableSliceFlagSeparator keeps such values intact.
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
