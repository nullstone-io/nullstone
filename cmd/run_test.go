package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadRunPayload(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "payload.json")
	require.NoError(t, os.WriteFile(filename, []byte(`{"from":"file"}`), 0644))

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name: "no payload",
			raw:  "",
			want: "",
		},
		{
			name: "inline payload",
			raw:  `{"from":"flag"}`,
			want: `{"from":"flag"}`,
		},
		{
			name: "payload from a file",
			raw:  "@" + filename,
			want: `{"from":"file"}`,
		},
		{
			name:    "payload from a missing file",
			raw:     "@" + filepath.Join(dir, "missing.json"),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := readRunPayload(test.raw)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, string(got))
		})
	}
}
