package runs

import (
	"encoding/json"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strconv"
	"strings"
)

func SetRunConfigVars(rc *types.RunConfig, varFlags []string) ([]string, error) {
	var errs []string
	skipped := make([]string, 0)

	for _, varFlag := range varFlags {
		tokens := strings.SplitN(varFlag, "=", 2)
		if len(tokens) < 2 {
			// We skip any variables that don't have an `=` sign
			continue
		}
		name, value := tokens[0], tokens[1]
		if v, ok := rc.Variables[name]; ok {
			if out, err := parseVarFlag(v, name, value); err != nil {
				errs = append(errs, err.Error())
			} else {
				v.Value = out
				rc.Variables[name] = v
			}
		} else {
			skipped = append(skipped, name)
		}
	}

	if len(errs) > 0 {
		return skipped, fmt.Errorf(`"--var" flags contain invalid values:
    * %s
`, strings.Join(errs, `
    * `))
	}
	return skipped, nil
}

func parseVarFlag(variable types.Variable, name, value string) (interface{}, error) {
	// Look in RunConfig for variable matching `name`
	switch variable.Type {
	case "string":
		return value, nil
	case "number":
		if iout, err := strconv.Atoi(value); err == nil {
			return iout, nil
		} else if fout, err := strconv.ParseFloat(value, 64); err == nil {
			return fout, nil
		} else {
			return nil, fmt.Errorf("%s: expected 'number' - %s", name, err)
		}
	case "bool":
		if out, err := strconv.ParseBool(value); err != nil {
			return nil, fmt.Errorf("%s: expected 'bool' - %s", name, err)
		} else {
			return out, nil
		}
	}

	var out interface{}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, fmt.Errorf("%s: expected json %s", name, err)
	}
	return out, nil
}
