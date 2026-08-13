package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// EnvSetTagFlag sets a tag on an environment. This is the write-side counterpart to
// EnvTagFilterFlag and the empty value means something different: `--tag claim=` sets
// claim to the empty string, which the API keeps distinct from an absent key. To remove
// a key entirely, use --remove-tag.
var EnvSetTagFlag = &cli.StringSliceFlag{
	Name: "tag",
	Usage: `Set a tag on the environment in the form KEY=VALUE.
		Can be specified multiple times. An empty value (--tag claim=) sets the tag to an empty
		string, which is not the same as removing it -- use --remove-tag for that.`,
}

var EnvRemoveTagFlag = &cli.StringSliceFlag{
	Name: "remove-tag",
	Usage: `Remove a tag from the environment by key. Can be specified multiple times.
		Removing a key that isn't set is not an error.`,
}

// ParseEnvTagSets parses the repeated --tag KEY=VALUE flag into a map, using the same
// parsing and error wording as --env-var.
func ParseEnvTagSets(c *cli.Context) (map[string]string, error) {
	raw := c.StringSlice(EnvSetTagFlag.Name)
	if len(raw) == 0 {
		return nil, nil
	}

	tags := make(map[string]string, len(raw))
	for _, kvp := range raw {
		tokens := strings.SplitN(kvp, "=", 2)
		if len(tokens) < 2 || tokens[0] == "" {
			return nil, fmt.Errorf("invalid --tag %q: must be in the form KEY=VALUE", kvp)
		}
		tags[tokens[0]] = tokens[1]
	}
	return tags, nil
}

// ParseEnvTagPatch combines --tag and --remove-tag into the per-key patch the API takes:
// a non-nil value sets the key, a nil value clears it, and an absent key is left alone.
// Returns nil when neither flag was given, so callers can tell "no tag changes" from
// "an empty set of changes".
func ParseEnvTagPatch(c *cli.Context) (map[string]*string, error) {
	sets, err := ParseEnvTagSets(c)
	if err != nil {
		return nil, err
	}
	removes := c.StringSlice(EnvRemoveTagFlag.Name)
	if len(sets) == 0 && len(removes) == 0 {
		return nil, nil
	}

	patch := make(map[string]*string, len(sets)+len(removes))
	for key, value := range sets {
		patch[key] = &value
	}
	for _, key := range removes {
		if key == "" {
			return nil, fmt.Errorf("invalid --remove-tag: a tag key is required")
		}
		if strings.Contains(key, "=") {
			return nil, fmt.Errorf("invalid --remove-tag %q: give the key only, not KEY=VALUE", key)
		}
		// Setting and removing the same key is a contradiction, not a precedence question.
		if _, ok := patch[key]; ok {
			return nil, fmt.Errorf("tag %q is both set and removed", key)
		}
		patch[key] = nil
	}
	return patch, nil
}
