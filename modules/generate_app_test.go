package modules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestGenerate_App(t *testing.T) {
	t.Chdir(t.TempDir())

	require.NoError(t, Generate(&types.ModuleManifest{Category: string(types.CategoryApp)}))

	scaffold, err := os.ReadFile(scaffoldTfFilename)
	require.NoError(t, err)
	assert.Contains(t, string(scaffold), `version = "~> 0.13.0"`)

	envVars, err := os.ReadFile(appEnvVarsTfFilename)
	require.NoError(t, err)
	for _, want := range []string{
		`data "ns_env_layout" "this"`,
		`data "ns_env_values" "this"`,
		`data "ns_env_platform_data" "this"`,
		`capability_prefixes    = local.cap_prefixes`,
	} {
		assert.Contains(t, string(envVars), want)
	}
	for _, unwanted := range []string{"ns_env_variables", "ns_secret_keys"} {
		assert.NotContains(t, string(envVars), unwanted)
	}
}
