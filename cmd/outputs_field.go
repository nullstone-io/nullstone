package cmd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// resolveOutputField resolves a Terraform-style traversal expression
// (e.g. `instance_id`, `endpoint.host`, `hosts[0]`, `hosts["item1"]`)
// against a set of workspace outputs.
func resolveOutputField(outputs types.Outputs, expr string) (any, error) {
	traversal, diags := hclsyntax.ParseTraversalAbs([]byte(expr), "--field", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("invalid --field expression %q: %s", expr, diags.Error())
	}

	rootName := traversal.RootName()
	output, ok := outputs[rootName]
	if !ok {
		return nil, fmt.Errorf("output %q not found (available outputs: %s)", rootName, strings.Join(outputNames(outputs), ", "))
	}
	if output.Redacted {
		return nil, fmt.Errorf("output %q is sensitive; re-run with --sensitive to emit its value", rootName)
	}

	cur := output.Value
	path := rootName
	for _, traverser := range traversal[1:] {
		var err error
		switch t := traverser.(type) {
		case hcl.TraverseAttr:
			cur, err = traverseKey(cur, t.Name, path)
			path = fmt.Sprintf("%s.%s", path, t.Name)
		case hcl.TraverseIndex:
			switch t.Key.Type() {
			case cty.String:
				key := t.Key.AsString()
				cur, err = traverseKey(cur, key, path)
				path = fmt.Sprintf("%s[%q]", path, key)
			case cty.Number:
				var idx int64
				if idx, err = ctyInt(t.Key); err == nil {
					cur, err = traverseIndex(cur, int(idx), path)
					path = fmt.Sprintf("%s[%d]", path, idx)
				}
			default:
				err = fmt.Errorf("%s: unsupported index type in --field expression", path)
			}
		default:
			err = fmt.Errorf("%s: unsupported traversal in --field expression", path)
		}
		if err != nil {
			return nil, err
		}
	}
	return cur, nil
}

// formatOutputValue renders a resolved output value for stdout: strings are
// emitted raw (unquoted) so they compose in command substitution; all other
// values are compact JSON.
func formatOutputValue(val any) (string, error) {
	if s, ok := val.(string); ok {
		return s, nil
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func traverseKey(val any, key string, path string) (any, error) {
	m, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s is not an object; cannot access %q", path, key)
	}
	child, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("%s has no key %q (available keys: %s)", path, key, strings.Join(mapKeys(m), ", "))
	}
	return child, nil
}

func traverseIndex(val any, idx int, path string) (any, error) {
	list, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("%s is not a list; cannot access index %d", path, idx)
	}
	if idx < 0 || idx >= len(list) {
		return nil, fmt.Errorf("%s[%d] is out of range (list has %d elements)", path, idx, len(list))
	}
	return list[idx], nil
}

func ctyInt(v cty.Value) (int64, error) {
	bf := v.AsBigFloat()
	idx, acc := bf.Int64()
	if acc != big.Exact {
		return 0, fmt.Errorf("index %s is not an integer", bf.String())
	}
	return idx, nil
}

func outputNames(outputs types.Outputs) []string {
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
