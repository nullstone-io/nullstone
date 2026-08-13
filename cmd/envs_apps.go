package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var EnvsApps = &cli.Command{
	Name: "apps",
	Description: `View and modify the apps configured in a preview environment.

In a preview environment an app is "enabled" by being present in the environment's preview app
set, so ` + "`--disabled`" + ` removes an app from the environment rather than writing a field.`,
	Usage:     "Manage the apps in a preview environment",
	UsageText: "nullstone envs apps [subcommand]",
	Subcommands: []*cli.Command{
		EnvsAppsList,
		EnvsAppsSet,
	},
}

var EnvsAppsList = &cli.Command{
	Name:        "list",
	Description: "Shows each app configured in the given preview environment, with the repo and the branch or pull request it tracks.",
	Usage:       "List the apps in a preview environment",
	UsageText:   "nullstone envs apps list --stack=<stack> --env=<env>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			stack, env, err := findPreviewEnv(ctx, cfg, c.String(StackRequiredFlag.Name), c.String(EnvFlag.Name))
			if err != nil {
				return err
			}

			previewApps, err := client.PreviewApps().List(ctx, stack.Id, env.Id)
			if err != nil {
				return fmt.Errorf("error listing preview apps: %w", err)
			}
			appNames, err := appNamesById(ctx, client, stack.Id)
			if err != nil {
				return err
			}

			sortPreviewApps(previewApps, appNames)
			rows := make([]string, len(previewApps)+1)
			rows[0] = "App|Repo|Tracking"
			for i, previewApp := range previewApps {
				rows[i+1] = fmt.Sprintf("%s|%s|%s", appNames[previewApp.AppId], previewApp.Repo, describeTracking(previewApp))
			}
			fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
			return nil
		})
	},
}

var EnvsAppsSet = &cli.Command{
	Name: "set",
	Description: `Updates every app already in the given preview environment that comes from --repo. Use --app
to narrow to a single app. Apps from other repos are left untouched, and a --repo that matches
nothing in the environment is an error rather than a no-op write.`,
	Usage:     "Configure the apps from a repo in a preview environment",
	UsageText: "nullstone envs apps set --stack=<stack> --env=<env> --repo=<owner/name> [--branch=<branch>|--pull-request=<number>|--disabled]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:     "repo",
			Usage:    "Select every app in this environment that comes from this repo, e.g. --repo=nullstone-io/nullstone",
			Required: true,
		},
		AppFlag,
		&cli.StringFlag{
			Name:  "branch",
			Usage: "Track this branch. Cannot be combined with --pull-request; setting it clears any pull request.",
		},
		&cli.Int64Flag{
			Name:  "pull-request",
			Usage: "Track this pull request. Cannot be combined with --branch; setting it clears any branch.",
		},
		&cli.BoolFlag{
			Name:  "disabled",
			Usage: "Remove the matched apps from this environment instead of updating them.",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			// Validate the flag combination before touching the API so a bad invocation
			// never results in a partial write.
			branch := c.String("branch")
			pullRequestId := c.Int64("pull-request")
			disabled := c.Bool("disabled")
			if c.IsSet("branch") && c.IsSet("pull-request") {
				return fmt.Errorf("--branch and --pull-request cannot be used together")
			}
			if disabled && (c.IsSet("branch") || c.IsSet("pull-request")) {
				return fmt.Errorf("--disabled cannot be combined with --branch or --pull-request")
			}
			if !disabled && !c.IsSet("branch") && !c.IsSet("pull-request") {
				return fmt.Errorf("specify one of --branch, --pull-request, or --disabled")
			}

			client := api.Client{Config: cfg}
			stack, env, err := findPreviewEnv(ctx, cfg, c.String(StackRequiredFlag.Name), c.String(EnvFlag.Name))
			if err != nil {
				return err
			}
			appNames, err := appNamesById(ctx, client, stack.Id)
			if err != nil {
				return err
			}

			// The API replaces the whole set, so read the current list and send it back
			// mutated. Anything not matched is carried through untouched.
			existing, err := client.PreviewApps().List(ctx, stack.Id, env.Id)
			if err != nil {
				return fmt.Errorf("error listing preview apps: %w", err)
			}

			update := previewAppUpdate{
				Repo:     c.String("repo"),
				AppName:  c.String(AppFlag.Name),
				Disabled: disabled,
			}
			if c.IsSet("branch") {
				update.BranchName = &branch
			}
			if c.IsSet("pull-request") {
				update.PullRequestId = &pullRequestId
			}

			updated, matched := update.ApplyTo(existing, appNames)
			if matched == 0 {
				return fmt.Errorf("no apps in environment %q come from repo %q%s", env.Name, update.Repo, narrowedByApp(update.AppName))
			}
			repo := update.Repo

			result, err := client.PreviewApps().Replace(ctx, stack.Id, env.Id, updated)
			if err != nil {
				return fmt.Errorf("error updating preview apps: %w", err)
			}

			verb := "updated"
			if disabled {
				verb = "removed"
			}
			fmt.Fprintf(os.Stderr, "%s %d app(s) from %q in environment %q\n", verb, matched, repo, env.Name)
			sortPreviewApps(result, appNames)
			for _, previewApp := range result {
				fmt.Println(appNames[previewApp.AppId])
			}
			return nil
		})
	},
}

// findPreviewEnv resolves the stack and env, rejecting anything that isn't a preview env
// because preview apps only exist there.
func findPreviewEnv(ctx context.Context, cfg api.Config, stackName, envName string) (*types.Stack, *types.Environment, error) {
	stack, err := find.Stack(ctx, cfg, stackName)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving stack: %w", err)
	} else if stack == nil {
		return nil, nil, fmt.Errorf("stack %q does not exist", stackName)
	}

	env, err := find.Env(ctx, cfg, stack.Id, envName)
	if err != nil {
		return nil, nil, fmt.Errorf("error looking for environment %q in stack %q: %w", envName, stack.Name, err)
	} else if env == nil {
		return nil, nil, fmt.Errorf("environment %q does not exist in stack %q", envName, stack.Name)
	}
	if env.Type != types.EnvTypePreview {
		return nil, nil, fmt.Errorf("environment %q is a %s environment; preview apps are only configured on preview environments",
			env.Name, strings.TrimSuffix(string(env.Type), "Env"))
	}
	return stack, env, nil
}

func appNamesById(ctx context.Context, client api.Client, stackId int64) (map[int64]string, error) {
	apps, err := client.Apps().List(ctx, stackId)
	if err != nil {
		return nil, fmt.Errorf("error listing apps: %w", err)
	}
	names := make(map[int64]string, len(apps))
	for _, app := range apps {
		names[app.Id] = app.Name
	}
	return names, nil
}

// previewAppUpdate is one `envs apps set` invocation: which apps it selects and what it
// does to them.
type previewAppUpdate struct {
	// Repo selects every preview app from this repo.
	Repo string
	// AppName narrows the selection to a single app. Empty selects them all.
	AppName string
	// Disabled removes the selected apps from the environment instead of updating them.
	Disabled bool
	// BranchName and PullRequestId are mutually exclusive; at most one is non-nil.
	BranchName    *string
	PullRequestId *int64
}

// ApplyTo returns the environment's full preview app set with the update applied, along
// with how many apps it matched. The API replaces the whole set, so unmatched apps are
// carried through verbatim — dropping one here would remove it from the environment.
func (u previewAppUpdate) ApplyTo(existing []types.PreviewApp, appNames map[int64]string) ([]types.PreviewApp, int) {
	updated := make([]types.PreviewApp, 0, len(existing))
	matched := 0
	for _, previewApp := range existing {
		if !u.matches(previewApp, appNames) {
			updated = append(updated, previewApp)
			continue
		}
		matched++
		if u.Disabled {
			continue
		}
		// BranchName and PullRequestId are mutually exclusive on the model, so setting
		// one always clears the other.
		previewApp.BranchName = u.BranchName
		previewApp.PullRequestId = u.PullRequestId
		updated = append(updated, previewApp)
	}
	return updated, matched
}

// matches reports whether a preview app is selected by --repo, optionally narrowed by
// --app. Repo comparison is case-insensitive because git hosts treat owner/name that way.
func (u previewAppUpdate) matches(previewApp types.PreviewApp, appNames map[int64]string) bool {
	if !strings.EqualFold(previewApp.Repo, u.Repo) {
		return false
	}
	return u.AppName == "" || strings.EqualFold(appNames[previewApp.AppId], u.AppName)
}

func describeTracking(previewApp types.PreviewApp) string {
	switch {
	case previewApp.BranchName != nil:
		return fmt.Sprintf("branch %s", *previewApp.BranchName)
	case previewApp.PullRequestId != nil:
		return fmt.Sprintf("pull request %d", *previewApp.PullRequestId)
	default:
		return "-"
	}
}

func narrowedByApp(appName string) string {
	if appName == "" {
		return ""
	}
	return fmt.Sprintf(" (narrowed to app %q)", appName)
}

func sortPreviewApps(previewApps []types.PreviewApp, appNames map[int64]string) {
	sort.SliceStable(previewApps, func(i, j int) bool {
		return appNames[previewApps[i].AppId] < appNames[previewApps[j].AppId]
	})
}
