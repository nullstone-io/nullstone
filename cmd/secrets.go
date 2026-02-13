package cmd

import (
	"context"
	"fmt"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var Secrets = &cli.Command{
	Name:      "secrets",
	Usage:     "View and modify secrets",
	UsageText: "nullstone secrets [subcommand]",
	Subcommands: []*cli.Command{
		SecretsList,
		SecretsCreate,
		SecretsUpdate,
	},
}

var SecretsList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of secrets in the cloud platform that is configured for the given stack and environment.",
	Usage:       "List secrets",
	UsageText:   "nullstone secrets list --stack=<stack-name> --env=<env-name>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			stackName := c.String(StackRequiredFlag.Name)
			envName := c.String(EnvFlag.Name)

			stack, err := find.Stack(ctx, cfg, stackName)
			if err != nil {
				return fmt.Errorf("error retrieving stack: %w", err)
			} else if stack == nil {
				return fmt.Errorf("stack %s does not exist", stackName)
			}

			env, err := find.Env(ctx, cfg, stack.Id, envName)
			if err != nil {
				return fmt.Errorf("error retrieving environment: %w", err)
			} else if env == nil {
				return fmt.Errorf("environment %s does not exist in stack %s", envName, stackName)
			}

			secrets, err := client.Secrets().List(ctx, stack.Id, env.Id, types.SecretLocation{})
			if err != nil {
				return fmt.Errorf("error listing secrets: %w", err)
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets found")
				return nil
			}

			secretDetails := make([]string, len(secrets)+1)
			secretDetails[0] = "Name|Platform|Id"
			for i, secret := range secrets {
				secretDetails[i+1] = fmt.Sprintf("%s|%s|%s", secret.Identity.Name, secret.Identity.Platform, secret.Identity.Id())
			}
			fmt.Println(columnize.Format(secretDetails, columnize.DefaultConfig()))

			return nil
		})
	},
}

var SecretsCreate = &cli.Command{
	Name:        "create",
	Description: "Creates a new secret in the given stack and environment.",
	Usage:       "Create a new secret",
	UsageText:   "nullstone secrets create --stack=<stack-name> --env=<env-name> --name=<secret-name> --value=<secret-value>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the secret to create",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "value",
			Usage:    "The value of the secret",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			stackName := c.String(StackRequiredFlag.Name)
			envName := c.String(EnvFlag.Name)
			secretName := c.String("name")
			secretValue := c.String("value")

			stack, err := find.Stack(ctx, cfg, stackName)
			if err != nil {
				return fmt.Errorf("error retrieving stack: %w", err)
			} else if stack == nil {
				return fmt.Errorf("stack %s does not exist", stackName)
			}

			env, err := find.Env(ctx, cfg, stack.Id, envName)
			if err != nil {
				return fmt.Errorf("error retrieving environment: %w", err)
			} else if env == nil {
				return fmt.Errorf("environment %s does not exist in stack %s", envName, stackName)
			}

			input := api.AddSecretInput{
				Identity: types.SecretIdentity{Name: secretName},
				Value:    secretValue,
			}
			secret, err := client.Secrets().Add(ctx, stack.Id, env.Id, input)
			if err != nil {
				return fmt.Errorf("error creating secret: %w", err)
			}

			fmt.Printf("created secret %q\n", secret.Identity.Id())
			return nil
		})
	},
}

var SecretsUpdate = &cli.Command{
	Name:        "update",
	Description: "Updates an existing secret in the given stack and environment.",
	Usage:       "Update a secret",
	UsageText:   "nullstone secrets update --stack=<stack-name> --env=<env-name> --name=<secret-name> --value=<secret-value>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the secret to update",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "value",
			Usage:    "The new value of the secret",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			stackName := c.String(StackRequiredFlag.Name)
			envName := c.String(EnvFlag.Name)
			secretName := c.String("name")
			secretValue := c.String("value")

			stack, err := find.Stack(ctx, cfg, stackName)
			if err != nil {
				return fmt.Errorf("error retrieving stack: %w", err)
			} else if stack == nil {
				return fmt.Errorf("stack %s does not exist", stackName)
			}

			env, err := find.Env(ctx, cfg, stack.Id, envName)
			if err != nil {
				return fmt.Errorf("error retrieving environment: %w", err)
			} else if env == nil {
				return fmt.Errorf("environment %s does not exist in stack %s", envName, stackName)
			}

			input := api.UpdateSecretInput{
				Value: secretValue,
			}
			secret, err := client.Secrets().Update(ctx, stack.Id, env.Id, secretName, input)
			if err != nil {
				return fmt.Errorf("error updating secret: %w", err)
			}

			fmt.Printf("updated secret %q\n", secret.Identity.Id())
			return nil
		})
	},
}
