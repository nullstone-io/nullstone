package artifacts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindLatestVersionSequence(t *testing.T) {
	t.Run("no artifacts", func(t *testing.T) {
		artifacts := []string{}
		sequence := FindLatestVersionSequence("8b0ce41", artifacts)
		assert.Equal(t, -1, sequence)
	})
	t.Run("one artifact", func(t *testing.T) {
		artifacts := []string{"8b0ce41"}
		sequence := FindLatestVersionSequence("8b0ce41", artifacts)
		assert.Equal(t, 0, sequence)
	})
	t.Run("multiple sequences", func(t *testing.T) {
		artifacts := []string{"8b0ce41", "8b0ce41-1", "8b0ce41-2"}
		sequence := FindLatestVersionSequence("8b0ce41", artifacts)
		assert.Equal(t, 2, sequence)
	})
	t.Run("multiple artifact roots, multiple sequences", func(t *testing.T) {
		artifacts := []string{"8b0ce41", "8b0ce41-1", "aafb1de-1", "aafb1de-2", "aafb1de-3"}
		sequence := FindLatestVersionSequence("8b0ce41", artifacts)
		assert.Equal(t, 1, sequence)
	})
}
