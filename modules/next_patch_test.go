package modules

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBumpPatch(t *testing.T) {
	tests := []struct {
		latest string
		want   string
	}{
		{
			latest: "1",
			want:   "1.0.1",
		},
		{
			latest: "1.0",
			want:   "1.0.1",
		},
		{
			latest: "1.0.0",
			want:   "1.0.1",
		},
		{
			latest: "1.2.0",
			want:   "1.2.1",
		},
		{
			latest: "1.2.0-pre",
			want:   "1.2.1",
		},
		{
			latest: "1.2.0-pre+build",
			want:   "1.2.1",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := BumpPatch(test.latest)
			got = strings.TrimPrefix(got, "v")
			assert.Equal(t, test.want, got)
		})
	}
}
