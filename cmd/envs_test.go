package cmd

import (
	"strings"
	"testing"

	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestParseEnvType(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    types.EnvironmentType
		wantErr string
	}{
		{name: "preview", raw: "preview", want: types.EnvTypePreview},
		{name: "pipeline", raw: "pipeline", want: types.EnvTypePipeline},
		{name: "global", raw: "global", want: types.EnvTypeGlobal},
		{name: "previews-shared", raw: "previews-shared", want: types.EnvTypePreviewsShared},
		{name: "mixed case", raw: "Preview", want: types.EnvTypePreview},
		{name: "surrounding whitespace", raw: "  preview  ", want: types.EnvTypePreview},
		{name: "unknown type", raw: "bogus", wantErr: `invalid --type "bogus"`},
		{name: "api value is not accepted", raw: "PreviewEnv", wantErr: `invalid --type "PreviewEnv"`},
		{name: "empty", raw: "", wantErr: `invalid --type ""`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseEnvType(test.raw)
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("parseEnvType(%q) expected an error, got %q", test.raw, got)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("parseEnvType(%q) error = %q, want it to contain %q", test.raw, err.Error(), test.wantErr)
				}
				// The error needs to tell the user what the valid options are
				for _, name := range envTypeNames() {
					if !strings.Contains(err.Error(), name) {
						t.Errorf("parseEnvType(%q) error = %q, want it to list valid option %q", test.raw, err.Error(), name)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEnvType(%q) unexpected error: %v", test.raw, err)
			}
			if got != test.want {
				t.Fatalf("parseEnvType(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

// Every environment type the API defines should be reachable through the --type flag
func TestEnvTypesByNameCoversAllEnvironmentTypes(t *testing.T) {
	allEnvTypes := []types.EnvironmentType{
		types.EnvTypePipeline,
		types.EnvTypePreview,
		types.EnvTypePreviewsShared,
		types.EnvTypeGlobal,
	}

	for _, envType := range allEnvTypes {
		found := false
		for _, mapped := range envTypesByName {
			if mapped == envType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("environment type %q is not reachable through --type", envType)
		}
	}
}
