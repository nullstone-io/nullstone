package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Issue #673: slice-flag comma-splitting corrupts KEY=VALUE flags like --env-var
// whose values contain commas (e.g. JSON payloads). This guards the app-level
// setting that cmd's flag tests assume.
func TestBuildDisablesSliceFlagSeparator(t *testing.T) {
	assert.True(t, Build().DisableSliceFlagSeparator)
}
