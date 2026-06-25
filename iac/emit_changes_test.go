package iac

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func renderValueDiff(prev, next any) string {
	var buf bytes.Buffer
	emitVariableValueDiff(&buf, "", prev, next)
	return ansiRe.ReplaceAllString(buf.String(), "")
}

func TestEmitVariableValueDiff(t *testing.T) {
	t.Run("scalar", func(t *testing.T) {
		assert.Equal(t, "256 => 512\n", renderValueDiff(256, 512))
	})

	t.Run("nested map shows only changed/added/removed items", func(t *testing.T) {
		prev := map[string]any{
			"connection_draining": map[string]any{"timeout_sec": float64(60)},
			"timeout_sec":         float64(600),
			"removed_key":         "gone",
		}
		next := map[string]any{
			"connection_draining": map[string]any{"timeout_sec": float64(60)},
			"timeout_sec":         300,
			"added_key":           "new",
		}
		want := "{\n" +
			"    + added_key: new\n" +
			"      connection_draining: {timeout_sec: 60}\n" +
			"    - removed_key: gone\n" +
			"    ~ timeout_sec: 600 => 300\n" +
			"}\n"
		assert.Equal(t, want, renderValueDiff(prev, next))
	})

	t.Run("nested map recurses into changed sub-map", func(t *testing.T) {
		prev := map[string]any{"connection_draining": map[string]any{"timeout_sec": float64(60), "enabled": true}}
		next := map[string]any{"connection_draining": map[string]any{"timeout_sec": 90, "enabled": true}}
		want := "{\n" +
			"    ~ connection_draining: {\n" +
			"          enabled: true\n" +
			"        ~ timeout_sec: 60 => 90\n" +
			"    }\n" +
			"}\n"
		assert.Equal(t, want, renderValueDiff(prev, next))
	})

	t.Run("list shows per-item add/remove/change", func(t *testing.T) {
		prev := []any{"a", "b", "c"}
		next := []any{"a", "x", "c", "d"}
		want := "[\n" +
			"      a\n" +
			"    ~ b => x\n" +
			"      c\n" +
			"    + d\n" +
			"]\n"
		assert.Equal(t, want, renderValueDiff(prev, next))
	})
}

func TestVariableValToString(t *testing.T) {
	tests := map[string]struct {
		val  any
		want string
	}{
		"nil":     {val: nil, want: ""},
		"string":  {val: "hello", want: "hello"},
		"bool":    {val: true, want: "true"},
		"int":     {val: 60, want: "60"},
		"int64":   {val: int64(600), want: "600"},
		"float64": {val: float64(60), want: "60"},
		// JSON-decoded numbers are float64; they must render as plain integers,
		// not as Go's "%!s(float64=60)" fallback.
		"nested object with float numbers": {
			val: map[string]any{
				"connection_draining": map[string]any{"timeout_sec": float64(60)},
				"timeout_sec":         float64(600),
			},
			want: "{connection_draining: {timeout_sec: 60}, timeout_sec: 600}",
		},
		"list of numbers": {
			val:  []any{float64(1), float64(2), float64(3)},
			want: "[1, 2, 3]",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, variableValToString(test.val))
		})
	}
}
