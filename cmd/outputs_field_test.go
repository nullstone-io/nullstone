package cmd

import (
	"strings"
	"testing"

	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestResolveOutputField(t *testing.T) {
	outputs := types.Outputs{
		"instance_id": types.Output{Type: "string", Value: "i-0abc123"},
		"port":        types.Output{Type: "number", Value: float64(5432)},
		"enabled":     types.Output{Type: "bool", Value: true},
		"hosts": types.Output{
			Type:  []any{"list", "string"},
			Value: []any{"host-a", "host-b"},
		},
		"endpoint": types.Output{
			Type:  "object",
			Value: map[string]any{"host": "db.example.com", "port": float64(5432)},
		},
		"items": types.Output{
			Type:  "map",
			Value: map[string]any{"item1": map[string]any{"id": "abc"}},
		},
		"db_password": types.Output{Type: "string", Value: "(hidden)", Sensitive: true, Redacted: true},
	}

	tests := []struct {
		name    string
		expr    string
		want    any
		wantErr string
	}{
		{name: "top-level string", expr: "instance_id", want: "i-0abc123"},
		{name: "top-level number", expr: "port", want: float64(5432)},
		{name: "top-level bool", expr: "enabled", want: true},
		{name: "list index", expr: "hosts[0]", want: "host-a"},
		{name: "attr access", expr: "endpoint.host", want: "db.example.com"},
		{name: "quoted map key", expr: `items["item1"]`, want: map[string]any{"id": "abc"}},
		{name: "quoted key then attr", expr: `items["item1"].id`, want: "abc"},
		{name: "unknown output", expr: "missing", wantErr: `output "missing" not found`},
		{name: "unknown key", expr: "endpoint.missing", wantErr: `endpoint has no key "missing"`},
		{name: "index out of range", expr: "hosts[5]", wantErr: "hosts[5] is out of range"},
		{name: "index into scalar", expr: "instance_id[0]", wantErr: "instance_id is not a list"},
		{name: "attr on scalar", expr: "instance_id.foo", wantErr: "instance_id is not an object"},
		{name: "redacted output", expr: "db_password", wantErr: `output "db_password" is sensitive`},
		{name: "invalid expression", expr: "hosts[", wantErr: "invalid --field expression"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveOutputField(outputs, tt.expr)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got value %v", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			gotStr, _ := formatOutputValue(got)
			wantStr, _ := formatOutputValue(tt.want)
			if gotStr != wantStr {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatOutputValue(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{name: "string is raw", val: "i-0abc123", want: "i-0abc123"},
		{name: "string with spaces is unquoted", val: "hello world", want: "hello world"},
		{name: "number", val: float64(5432), want: "5432"},
		{name: "bool", val: true, want: "true"},
		{name: "nil", val: nil, want: "null"},
		{name: "list is compact json", val: []any{"a", "b"}, want: `["a","b"]`},
		{name: "map is compact json", val: map[string]any{"host": "db", "port": float64(1)}, want: `{"host":"db","port":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatOutputValue(tt.val)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
