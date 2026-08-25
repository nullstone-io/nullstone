package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestParseContributors(t *testing.T) {
	t.Run("repeated flags", func(t *testing.T) {
		contributors, err := parseContributors([]string{"nullstone-official", "my-org"})
		require.NoError(t, err)
		assert.Equal(t, []types.Contributor{types.ContributorNullstoneOfficial, types.ContributorMyOrg}, contributors)
	})

	// The app disables urfave/cli's comma-splitting (issue #673), so the
	// comma-separated form documented in the usage text is handled here.
	t.Run("comma-separated in a single flag", func(t *testing.T) {
		contributors, err := parseContributors([]string{"nullstone-official,my-org"})
		require.NoError(t, err)
		assert.Equal(t, []types.Contributor{types.ContributorNullstoneOfficial, types.ContributorMyOrg}, contributors)
	})

	t.Run("invalid value errors", func(t *testing.T) {
		_, err := parseContributors([]string{"nullstone-official,bogus"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `invalid --contributor "bogus"`)
	})
}
