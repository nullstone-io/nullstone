package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHashArchiveContents locks the content-digest algorithm. The pinned hex MUST
// match the equivalent test in oracle (manifest/scan_artifact_test.go) — if the
// two ever diverge, --if=checksum-changed comparisons will fail on first run.
func TestHashArchiveContents(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("test-fixtures", "block-aws-network.tgz"))
	require.NoError(t, err, "read test fixture")

	got, err := HashArchiveContents(data, ".tgz")
	require.NoError(t, err, "hash archive")
	require.NotEmpty(t, got)

	// Deterministic: identical bytes produce identical checksums.
	got2, err := HashArchiveContents(data, ".tgz")
	require.NoError(t, err)
	assert.Equal(t, got, got2, "checksum must be deterministic")

	// Pinned: locks both the algorithm and the fixture. If the algorithm changes,
	// update this literal AND the matching one in
	// oracle/manifest/scan_artifact_test.go.
	const expected = "786253ef007b45958a723fdd115e8b7fb68b1f6b4f4627cac1636c4aaaf8878d"
	assert.Equal(t, expected, got)

	// Sensitivity: mutating one byte changes the checksum.
	mutated := make([]byte, len(data))
	copy(mutated, data)
	mutated[len(mutated)/2] ^= 0xFF
	if got3, err := HashArchiveContents(mutated, ".tgz"); err == nil {
		assert.NotEqual(t, got, got3)
	}
}
