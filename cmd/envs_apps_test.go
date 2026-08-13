package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func strPtr(s string) *string { return &s }
func i64Ptr(i int64) *int64   { return &i }

var previewAppNames = map[int64]string{1: "api", 2: "worker", 3: "web"}

func previewAppFixtures() []types.PreviewApp {
	return []types.PreviewApp{
		{AppId: 1, Repo: "nullstone-io/backend", BranchName: strPtr("main")},
		{AppId: 2, Repo: "nullstone-io/backend", PullRequestId: i64Ptr(42)},
		{AppId: 3, Repo: "nullstone-io/frontend", BranchName: strPtr("main")},
	}
}

func TestPreviewAppUpdate_ApplyTo_Branch(t *testing.T) {
	update := previewAppUpdate{Repo: "nullstone-io/backend", BranchName: strPtr("feat/x")}

	updated, matched := update.ApplyTo(previewAppFixtures(), previewAppNames)

	assert.Equal(t, 2, matched)
	require.Len(t, updated, 3)

	// Every app from the repo now tracks the branch, and the pull request is cleared.
	assert.Equal(t, strPtr("feat/x"), updated[0].BranchName)
	assert.Nil(t, updated[0].PullRequestId)
	assert.Equal(t, strPtr("feat/x"), updated[1].BranchName)
	assert.Nil(t, updated[1].PullRequestId)

	// The app from the other repo is untouched and still present.
	assert.Equal(t, types.PreviewApp{AppId: 3, Repo: "nullstone-io/frontend", BranchName: strPtr("main")}, updated[2])
}

func TestPreviewAppUpdate_ApplyTo_PullRequest(t *testing.T) {
	update := previewAppUpdate{Repo: "nullstone-io/backend", PullRequestId: i64Ptr(99)}

	updated, matched := update.ApplyTo(previewAppFixtures(), previewAppNames)

	assert.Equal(t, 2, matched)
	// Setting a pull request clears the branch.
	assert.Nil(t, updated[0].BranchName)
	assert.Equal(t, i64Ptr(99), updated[0].PullRequestId)
	assert.Nil(t, updated[1].BranchName)
	assert.Equal(t, i64Ptr(99), updated[1].PullRequestId)
}

func TestPreviewAppUpdate_ApplyTo_Disabled(t *testing.T) {
	update := previewAppUpdate{Repo: "nullstone-io/backend", Disabled: true}

	updated, matched := update.ApplyTo(previewAppFixtures(), previewAppNames)

	assert.Equal(t, 2, matched)
	// Disabling removes the matched apps from the set and leaves everything else.
	require.Len(t, updated, 1)
	assert.Equal(t, int64(3), updated[0].AppId)
}

func TestPreviewAppUpdate_ApplyTo_NarrowedByApp(t *testing.T) {
	update := previewAppUpdate{Repo: "nullstone-io/backend", AppName: "worker", BranchName: strPtr("feat/x")}

	updated, matched := update.ApplyTo(previewAppFixtures(), previewAppNames)

	assert.Equal(t, 1, matched)
	require.Len(t, updated, 3)
	// Only the named app changed; its sibling from the same repo is untouched.
	assert.Equal(t, strPtr("main"), updated[0].BranchName)
	assert.Equal(t, strPtr("feat/x"), updated[1].BranchName)
	assert.Nil(t, updated[1].PullRequestId)
}

func TestPreviewAppUpdate_ApplyTo_NoMatches(t *testing.T) {
	existing := previewAppFixtures()

	t.Run("unknown repo matches nothing and changes nothing", func(t *testing.T) {
		update := previewAppUpdate{Repo: "nullstone-io/missing", BranchName: strPtr("feat/x")}

		updated, matched := update.ApplyTo(existing, previewAppNames)

		assert.Equal(t, 0, matched)
		assert.Equal(t, existing, updated)
	})

	t.Run("app not in the repo matches nothing", func(t *testing.T) {
		update := previewAppUpdate{Repo: "nullstone-io/backend", AppName: "web", BranchName: strPtr("feat/x")}

		_, matched := update.ApplyTo(existing, previewAppNames)

		assert.Equal(t, 0, matched)
	})
}

func TestPreviewAppUpdate_ApplyTo_IsCaseInsensitive(t *testing.T) {
	update := previewAppUpdate{Repo: "NullStone-IO/BackEnd", AppName: "WORKER", BranchName: strPtr("feat/x")}

	_, matched := update.ApplyTo(previewAppFixtures(), previewAppNames)

	assert.Equal(t, 1, matched)
}

func TestPreviewAppUpdate_ApplyTo_DoesNotMutateInput(t *testing.T) {
	existing := previewAppFixtures()
	update := previewAppUpdate{Repo: "nullstone-io/backend", BranchName: strPtr("feat/x")}

	update.ApplyTo(existing, previewAppNames)

	assert.Equal(t, previewAppFixtures(), existing)
}

func TestDescribeTracking(t *testing.T) {
	assert.Equal(t, "branch main", describeTracking(types.PreviewApp{BranchName: strPtr("main")}))
	assert.Equal(t, "pull request 42", describeTracking(types.PreviewApp{PullRequestId: i64Ptr(42)}))
	assert.Equal(t, "-", describeTracking(types.PreviewApp{}))
}
