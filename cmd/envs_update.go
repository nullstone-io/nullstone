package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var EnvsUpdate = &cli.Command{
	Name: "update",
	Description: `Updates an existing environment. Only the attributes you pass are changed; everything
else is left alone.

Tags are applied as a per-key patch, so --tag adds or updates a single key and --remove-tag
deletes one, both without disturbing the environment's other tags. Setting a tag to an empty
value (--tag claim=) keeps the key with an empty value, which is not the same as removing it.`,
	Usage:     "Update an environment",
	UsageText: "nullstone envs update --stack=<stack> --env=<env> [--name=<name>] [--description=<text>] [--tag KEY=VALUE] [--remove-tag KEY]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:  "name",
			Usage: "Rename the environment.",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Describe what this environment is for. Pass an empty value to clear it.",
		},
		&cli.BoolFlag{
			Name:  "prod",
			Usage: "Mark this environment as production. Cannot be combined with --non-prod.",
		},
		&cli.BoolFlag{
			Name:  "non-prod",
			Usage: "Mark this environment as non-production. Cannot be combined with --prod.",
		},
		EnvSetTagFlag,
		EnvRemoveTagFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			// Build and validate the whole request before touching the API, so a bad
			// invocation never lands a partial update.
			input, tagPatch, err := parseEnvUpdate(c)
			if err != nil {
				return err
			}

			client := api.Client{Config: cfg}
			stackName, envName := c.String(StackRequiredFlag.Name), c.String(EnvFlag.Name)
			stack, err := find.Stack(ctx, cfg, stackName)
			if err != nil {
				return fmt.Errorf("error retrieving stack: %w", err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist", stackName)
			}
			env, err := find.Env(ctx, cfg, stack.Id, envName)
			if err != nil {
				return fmt.Errorf("error looking for environment %q in stack %q: %w", envName, stack.Name, err)
			} else if env == nil {
				return fmt.Errorf("environment %q does not exist in stack %q", envName, stack.Name)
			}

			updated := env
			if input != nil {
				if updated, err = client.Environments().Update(ctx, stack.Id, env.Id, *input); err != nil {
					return fmt.Errorf("error updating environment: %w", err)
				}
			}
			// Tags are a separate atomic endpoint, so they're a separate call. It runs
			// second; if it fails the attribute update above has already landed, which the
			// error message says outright rather than implying nothing happened.
			if tagPatch != nil {
				updated, err = client.Environments().UpdateTags(ctx, stack.Id, env.Id, api.UpdateEnvironmentTagsInput{Tags: tagPatch})
				if err != nil {
					if input != nil {
						return fmt.Errorf("environment attributes were updated, but updating tags failed: %w", err)
					}
					return fmt.Errorf("error updating tags: %w", err)
				}
			}

			fmt.Fprintf(os.Stderr, "updated %q environment\n", updated.Name)
			printEnvDetail(updated)
			return nil
		})
	},
}

// parseEnvUpdate turns the flags into the two API payloads. A nil input means no
// attribute changed; a nil patch means no tag changed. Both nil is an error, since a
// no-op update is more likely a mistake than an intent.
func parseEnvUpdate(c *cli.Context) (*api.UpdateEnvironmentInput, map[string]*string, error) {
	tagPatch, err := ParseEnvTagPatch(c)
	if err != nil {
		return nil, nil, err
	}

	isProd, isNonProd := c.Bool("prod"), c.Bool("non-prod")
	if isProd && isNonProd {
		return nil, nil, fmt.Errorf("--prod and --non-prod cannot be used together")
	}

	var input *api.UpdateEnvironmentInput
	set := func() *api.UpdateEnvironmentInput {
		if input == nil {
			input = &api.UpdateEnvironmentInput{}
		}
		return input
	}
	if c.IsSet("name") {
		name := sanitizeEnvName(c.String("name"))
		set().Name = &name
	}
	// IsSet, not the value: an explicitly empty --description clears the description.
	if c.IsSet("description") {
		description := c.String("description")
		set().Metadata = &api.UpdateEnvironmentMetadataInput{Description: &description}
	}
	if isProd || isNonProd {
		set().IsProd = &isProd
	}

	if input == nil && tagPatch == nil {
		return nil, nil, fmt.Errorf("nothing to update: pass at least one of --name, --description, --prod, --non-prod, --tag, or --remove-tag")
	}
	return input, tagPatch, nil
}

func printEnvDetail(env *types.Environment) {
	rows := []string{
		fmt.Sprintf("Name|%s", env.Name),
		fmt.Sprintf("Type|%s", env.Type),
		fmt.Sprintf("Prod|%t", env.IsProd),
		fmt.Sprintf("Description|%s", env.Metadata.Description),
	}
	keys := make([]string, 0, len(env.Tags))
	for key := range env.Tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		rows = append(rows, fmt.Sprintf("Tag %s|%s", key, env.Tags[key]))
	}
	fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
}
