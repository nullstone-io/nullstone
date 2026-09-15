package workspaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigSource(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    ConfigSource
		wantErr string
	}{
		{name: "empty defaults to effective", raw: "", want: ConfigSourceEffective},
		{name: "effective", raw: "effective", want: ConfigSourceEffective},
		{name: "latest", raw: "latest", want: ConfigSourceLatest},
		{name: "current", raw: "current", want: ConfigSourceCurrent},
		{name: "case and whitespace insensitive", raw: "  Latest ", want: ConfigSourceLatest},
		{name: "stable is not exposed", raw: "stable", wantErr: `invalid --config "stable": must be one of effective, latest, current`},
		{name: "garbage", raw: "nope", wantErr: `invalid --config "nope": must be one of effective, latest, current`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfigSource(tt.raw)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
