package cmd

import (
	"strings"
	"testing"
)

func TestParsePullRequest(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantErr string
	}{
		{name: "pull request number", raw: "123", want: 123},
		{name: "pull request id", raw: "98765432", want: 98765432},
		{name: "leading hash", raw: "#123", want: 123},
		{name: "surrounding whitespace", raw: "  123  ", want: 123},
		{name: "not a number", raw: "abc", wantErr: `invalid --pull-request "abc"`},
		{name: "empty", raw: "", wantErr: `invalid --pull-request ""`},
		{name: "branch name", raw: "feat/some-branch", wantErr: `invalid --pull-request "feat/some-branch"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsePullRequest(test.raw)
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("parsePullRequest(%q) expected an error, got %d", test.raw, *got)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("parsePullRequest(%q) error = %q, want it to contain %q", test.raw, err.Error(), test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePullRequest(%q) unexpected error: %v", test.raw, err)
			}
			if got == nil {
				t.Fatalf("parsePullRequest(%q) = nil, want %d", test.raw, test.want)
			}
			if *got != test.want {
				t.Fatalf("parsePullRequest(%q) = %d, want %d", test.raw, *got, test.want)
			}
		})
	}
}
