package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		params   map[string]interface{}
		// wantPrefix contains the expected prefix of args (global flags + subcommands)
		wantPrefix []string
		// wantContains contains flag args that must appear somewhere after the prefix
		wantContains []string
	}{
		{
			name:       "simple list no params",
			toolName:   "stacks_list",
			params:     map[string]interface{}{},
			wantPrefix: []string{"stacks", "list"},
		},
		{
			name:         "list with detail flag",
			toolName:     "stacks_list",
			params:       map[string]interface{}{"detail": true},
			wantPrefix:   []string{"stacks", "list"},
			wantContains: []string{"--detail"},
		},
		{
			name:         "global flags before subcommand",
			toolName:     "stacks_list",
			params:       map[string]interface{}{"profile": "staging", "org": "myorg"},
			wantPrefix:   []string{"--profile", "staging", "--org", "myorg", "stacks", "list"},
			wantContains: nil,
		},
		{
			name:         "set_org positional arg",
			toolName:     "set_org",
			params:       map[string]interface{}{"org_name": "acme-corp"},
			wantPrefix:   []string{"set-org"},
			wantContains: []string{"acme-corp"},
		},
		{
			name:         "apply with multiple flags",
			toolName:     "apply",
			params:       map[string]interface{}{"block": "api", "env": "prod", "wait": true, "auto_approve": true},
			wantPrefix:   []string{"apply"},
			wantContains: []string{"--block", "api", "--env", "prod", "--wait", "--auto-approve"},
		},
		{
			name:         "underscore to hyphen conversion",
			toolName:     "plan",
			params:       map[string]interface{}{"block": "api", "env": "dev", "module_version": "1.2.3"},
			wantPrefix:   []string{"plan"},
			wantContains: []string{"--module-version", "1.2.3"},
		},
		{
			name:         "array params repeated",
			toolName:     "up",
			params:       map[string]interface{}{"block": "api", "env": "dev", "var": []interface{}{"key1=val1", "key2=val2"}},
			wantPrefix:   []string{"up"},
			wantContains: []string{"--var", "key1=val1", "--var", "key2=val2"},
		},
		{
			name:         "top-level command",
			toolName:     "outputs",
			params:       map[string]interface{}{"block": "web", "env": "staging"},
			wantPrefix:   []string{"outputs"},
			wantContains: []string{"--block", "web", "--env", "staging"},
		},
		{
			name:         "bool false not included",
			toolName:     "stacks_list",
			params:       map[string]interface{}{"detail": false},
			wantPrefix:   []string{"stacks", "list"},
			wantContains: nil,
		},
		{
			name:         "empty string not included",
			toolName:     "outputs",
			params:       map[string]interface{}{"block": "web", "env": "staging", "stack": ""},
			wantPrefix:   []string{"outputs"},
			wantContains: []string{"--block", "web", "--env", "staging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildArgs(tt.toolName, tt.params)

			// Check prefix (global flags + subcommands come first)
			if len(tt.wantPrefix) > 0 {
				assert.True(t, len(got) >= len(tt.wantPrefix),
					"args too short: got %v, want prefix %v", got, tt.wantPrefix)
				assert.Equal(t, tt.wantPrefix, got[:len(tt.wantPrefix)],
					"prefix mismatch")
			}

			// Check that all expected flag args are present
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want,
					"expected arg %q in %v", want, got)
			}

			// Empty strings should not appear
			for _, arg := range got {
				assert.NotEmpty(t, arg, "empty string in args")
			}
		})
	}
}

func TestToolNameToSubcommands(t *testing.T) {
	tests := []struct {
		toolName string
		want     []string
	}{
		{"stacks_list", []string{"stacks", "list"}},
		{"stacks_new", []string{"stacks", "new"}},
		{"envs_list", []string{"envs", "list"}},
		{"envs_new", []string{"envs", "new"}},
		{"envs_delete", []string{"envs", "delete"}},
		{"envs_up", []string{"envs", "up"}},
		{"envs_down", []string{"envs", "down"}},
		{"apps_list", []string{"apps", "list"}},
		{"blocks_list", []string{"blocks", "list"}},
		{"blocks_new", []string{"blocks", "new"}},
		{"set_org", []string{"set-org"}},
		{"workspaces_select", []string{"workspaces", "select"}},
		{"iac_test", []string{"iac", "test"}},
		{"iac_generate", []string{"iac", "generate"}},
		{"modules_register", []string{"modules", "register"}},
		{"modules_publish", []string{"modules", "publish"}},
		{"modules_package", []string{"modules", "package"}},
		// Top-level commands
		{"outputs", []string{"outputs"}},
		{"status", []string{"status"}},
		{"profile", []string{"profile"}},
		{"logs", []string{"logs"}},
		{"up", []string{"up"}},
		{"plan", []string{"plan"}},
		{"apply", []string{"apply"}},
		{"wait", []string{"wait"}},
		{"push", []string{"push"}},
		{"deploy", []string{"deploy"}},
		{"launch", []string{"launch"}},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := toolNameToSubcommands(tt.toolName)
			assert.Equal(t, tt.want, got)
		})
	}
}
