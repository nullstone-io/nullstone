package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strconv"
	"strings"
	"time"
)

var Up = func() *cli.Command {
	return &cli.Command{
		Name:      "up",
		Usage:     "Provisions the block and all of its dependencies",
		UsageText: "nullstone up [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for Nullstone to fully provision the workspace.",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variable values when issuing `up`",
			},
		},
		Action: func(c *cli.Context) error {
			return BlockEnvAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment) error {
				varFlags := c.StringSlice("var")

				client := api.Client{Config: cfg}
				workspace, err := client.Workspaces().Get(stack.Id, block.Id, env.Id)
				if err != nil {
					return fmt.Errorf("error looking for workspace: %w", err)
				} else if workspace == nil {
					return fmt.Errorf("workspace not found")
				}

				if workspace.Status == types.WorkspaceStatusProvisioned {
					fmt.Println("workspace is already provisioned")
					return nil
				}

				newRunConfig, err := client.PromotionConfigs().Get(workspace.StackId, workspace.BlockId, workspace.EnvId)
				if err != nil {
					return err
				}

				fillRunConfigVariables(newRunConfig)

				if err := setRunConfigVars(newRunConfig, varFlags); err != nil {
					return err
				}

				isApproved := true
				input := types.CreateRunInput{
					IsDestroy:         false,
					IsApproved:        &isApproved,
					Source:            newRunConfig.Source,
					SourceVersion:     newRunConfig.SourceVersion,
					Variables:         newRunConfig.Variables,
					EnvVariables:      newRunConfig.EnvVariables,
					Connections:       newRunConfig.Connections,
					Capabilities:      newRunConfig.Capabilities,
					Providers:         newRunConfig.Providers,
					DependencyConfigs: newRunConfig.DependencyConfigs,
				}

				newRun, err := client.Runs().Create(workspace.StackId, workspace.Uid, input)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Printf("created run %q\n", newRun.Uid)

				if c.IsSet("wait") {
					return streamLiveLogs(ctx, cfg, workspace, newRun)
				}

				return nil
			})
		},
	}
}

func streamLiveLogs(ctx context.Context, cfg api.Config, workspace *types.Workspace, newRun *types.Run) error {
	// ctx already contains cancellation for Ctrl+C
	// innerCtx will allow us to cancel when the run reaches a terminal status
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	fmt.Println("Wating for logs...")
	client := api.Client{Config: cfg}
	msgs, err := client.LiveLogs().Watch(innerCtx, workspace.StackId, newRun.Uid)
	if err != nil {
		return err
	}
	runCh := pollRun(innerCtx, cfg, workspace.StackId, newRun.Uid, time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.Source != "error" {
				fmt.Print(msg.Content)
			}
		case run := <-runCh:
			if types.IsTerminalRunStatus(run.Status) {
				cancelFn()
				return nil
			}
		}
	}
}

func pollRun(ctx context.Context, cfg api.Config, stackId int64, runUid uuid.UUID, pollDelay time.Duration) <-chan types.Run {
	ch := make(chan types.Run)
	client := api.Client{Config: cfg}
	go func() {
		defer close(ch)
		for {
			run, _ := client.Runs().Get(stackId, runUid)
			if run != nil {
				ch <- *run
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(pollDelay):
			}
		}
	}()
	return ch
}

func fillRunConfigVariables(rc *types.RunConfig) {
	rc.Variables = fillVariables(rc.Variables)
	for i, c := range rc.Capabilities {
		c.Variables = fillVariables(c.Variables)
		rc.Capabilities[i] = c
	}
	for i, dc := range rc.DependencyConfigs {
		dc.Variables = fillVariables(dc.Variables)
		rc.DependencyConfigs[i] = dc
	}
}

func fillVariables(vars types.Variables) types.Variables {
	for k, v := range vars {
		if v.Value == nil {
			v.Value = v.Default
		}
		vars[k] = v
	}
	return vars
}

func setRunConfigVars(rc *types.RunConfig, varFlags []string) error {
	var errs []string

	for _, varFlag := range varFlags {
		tokens := strings.SplitN(varFlag, "=", 2)
		if len(tokens) < 2 {
			// We skip any variables that don't have an `=` sign
			continue
		}
		name, value := tokens[0], tokens[1]
		// Look in RunConfig for variable matching `name`
		// If we don't find a matching variable, we just skip it
		if v, ok := rc.Variables[name]; ok {
			if out, err := parseVarFlag(v, name, value); err != nil {
				errs = append(errs, err.Error())
			} else {
				v.Value = out
				rc.Variables[name] = v
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(`"--var" flags contain invalid values:
    * %s
`, strings.Join(errs, `
    * `))
	}
	return nil
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
