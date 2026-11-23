package cmd

import (
	"context"
	"fmt"
	"io"
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
	Description: `View and modify the apps of a preview environment.

A preview environment follows one branch or pull request per repo, and each app is turned on or off
in the environment. ` + "`--branch`" + ` and ` + "`--pull-request`" + ` change what the environment follows for a whole repo;
` + "`--enabled`" + ` and ` + "`--disabled`" + ` turn apps on or off.`,
	Usage:     "Manage the apps in a preview environment",
	UsageText: "nullstone envs apps [subcommand]",
	Subcommands: []*cli.Command{
		EnvsAppsList,
		EnvsAppsSet,
	},
}

var EnvsAppsList = &cli.Command{
	Name:        "list",
	Description: "Shows each app that is enabled in the given preview environment, with the repo and the branch or pull request it tracks.",
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
	Description: `Configures the apps that come from --repo in the given preview environment.

--branch or --pull-request sets what the environment follows for the whole repo, so they cannot be
narrowed with --app. --enabled or --disabled turns on or off every app from the repo; use --app to
narrow to a single app. A --repo that matches no app in the stack is an error rather than a no-op.`,
	Usage:     "Configure the apps from a repo in a preview environment",
	UsageText: "nullstone envs apps set --stack=<stack> --env=<env> --repo=<owner/name> [--branch=<branch>|--pull-request=<id>|--enabled|--disabled]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:     "repo",
			Usage:    "Select every app in the stack that comes from this repo, e.g. --repo=nullstone-io/nullstone",
			Required: true,
		},
		AppFlag,
		&cli.StringFlag{
			Name:  "branch",
			Usage: "Follow this branch of the repo in this environment. Replaces a pull request the environment was following.",
		},
		&cli.Int64Flag{
			Name:  "pull-request",
			Usage: "Follow this pull request of the repo in this environment, given as its Nullstone pull request id. Replaces a branch the environment was following.",
		},
		&cli.BoolFlag{
			Name:  "enabled",
			Usage: "Turn on the matched apps in this environment.",
		},
		&cli.BoolFlag{
			Name:  "disabled",
			Usage: "Turn off the matched apps in this environment.",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			update := envAppsUpdate{
				Repo:    c.String("repo"),
				AppName: c.String(AppFlag.Name),
			}
			if c.IsSet("branch") {
				branch := c.String("branch")
				update.BranchName = &branch
			}
			if c.IsSet("pull-request") {
				pullRequestId := c.Int64("pull-request")
				update.PullRequestId = &pullRequestId
			}
			update.Enable = c.Bool("enabled")
			update.Disable = c.Bool("disabled")
			// Validate the flag combination before touching the API so a bad invocation
			// never results in a partial write.
			if err := update.Validate(); err != nil {
				return err
			}

			return setEnvApps(ctx, cfg, c.String(StackRequiredFlag.Name), c.String(EnvFlag.Name), update, os.Stdout, os.Stderr)
		})
	},
}

// setEnvApps runs a validated `envs apps set`. It reports what it did on stderr and prints the names of the
// apps it applied to on stdout, one per line.
func setEnvApps(ctx context.Context, cfg api.Config, stackName, envName string, update envAppsUpdate, stdout, stderr io.Writer) error {
	client := api.Client{Config: cfg}
	stack, env, err := findPreviewEnv(ctx, cfg, stackName, envName)
	if err != nil {
		return err
	}
	apps, err := client.Apps().List(ctx, stack.Id)
	if err != nil {
		return fmt.Errorf("error listing apps: %w", err)
	}

	// TEMPORARY: a server that still keeps a preview environment's apps as a list (see envs_apps_legacy.go)
	followsTriggers, err := envFollowsTriggers(ctx, &client, stack.Id, env.Id)
	if err != nil {
		return fmt.Errorf("error reading what environment %q follows: %w", env.Name, err)
	}
	var matched []types.Application
	if followsTriggers {
		matched, err = update.apply(ctx, &client, *stack, *env, apps)
	} else {
		matched, err = setLegacyEnvApps(ctx, &client, *stack, *env, apps, update)
	}
	if err != nil {
		return err
	}

	switch {
	case update.SetsTrigger():
		fmt.Fprintf(stderr, "environment %q now follows %s of %q\n", env.Name, update.describeTrigger(), update.Repo)
	case update.Disable:
		fmt.Fprintf(stderr, "disabled %d app(s) from %q in environment %q\n", len(matched), update.Repo, env.Name)
	default:
		fmt.Fprintf(stderr, "enabled %d app(s) from %q in environment %q\n", len(matched), update.Repo, env.Name)
	}
	for _, app := range matched {
		fmt.Fprintln(stdout, app.Name)
	}
	return nil
}

// apply writes the update: what the environment follows from the repo, or which of its apps are on.
// It returns the apps it applied to.
func (u envAppsUpdate) apply(ctx context.Context, client *api.Client, stack types.Stack, env types.Environment, apps []types.Application) ([]types.Application, error) {
	matched := u.Select(apps)
	if len(matched) == 0 {
		return nil, fmt.Errorf("no apps in stack %q come from repo %q%s", stack.Name, u.Repo, narrowedByApp(u.AppName))
	}

	if u.SetsTrigger() {
		trigger, err := u.Trigger(matched)
		if err != nil {
			return nil, err
		}
		configs, err := client.EnvGitopsConfigs().Replace(ctx, stack.Id, env.Id, []api.SetEnvGitopsTriggerInput{trigger})
		if err != nil {
			return nil, fmt.Errorf("error updating what environment %q follows: %w", env.Name, err)
		}
		// The API client reads a 404 as an empty result. The response always lists the repo that was just set,
		// so nothing in it means nothing was written.
		if len(configs) == 0 {
			return nil, fmt.Errorf("error updating what environment %q follows: the server did not accept the change", env.Name)
		}
		return matched, nil
	}

	// The API writes each automation as a whole, so read the current ones and send
	// them back with only the enabled flag changed.
	current, err := client.WorkspaceAutomations().ListByEnv(ctx, stack.Id, env.Id)
	if err != nil {
		return nil, fmt.Errorf("error listing the automation of environment %q: %w", env.Name, err)
	}
	if _, err := client.WorkspaceAutomations().ReplaceByEnv(ctx, stack.Id, env.Id, u.Automations(matched, current)); err != nil {
		return nil, fmt.Errorf("error updating apps in environment %q: %w", env.Name, err)
	}
	return matched, nil
}

// findPreviewEnv resolves the stack and env, rejecting anything that isn't a preview env
// because these commands only configure preview environments.
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

// envAppsUpdate is one `envs apps set` invocation: which apps it selects and what it does.
//
// A preview environment follows one branch or pull request per repo, so BranchName and
// PullRequestId apply to the whole repo. Enable and Disable apply to each selected app.
type envAppsUpdate struct {
	// Repo selects every app in the stack that comes from this repo.
	Repo string
	// AppName narrows the selection to a single app. Empty selects them all.
	AppName string
	// Exactly one of BranchName, PullRequestId, Enable and Disable is set.
	BranchName    *string
	PullRequestId *int64
	Enable        bool
	Disable       bool
}

// Validate rejects flag combinations that have no single meaning
func (u envAppsUpdate) Validate() error {
	if u.BranchName != nil && u.PullRequestId != nil {
		return fmt.Errorf("--branch and --pull-request cannot be used together")
	}
	if u.Enable && u.Disable {
		return fmt.Errorf("--enabled and --disabled cannot be used together")
	}
	if (u.Enable || u.Disable) && u.SetsTrigger() {
		return fmt.Errorf("--enabled and --disabled cannot be combined with --branch or --pull-request")
	}
	if !u.Enable && !u.Disable && !u.SetsTrigger() {
		return fmt.Errorf("specify one of --branch, --pull-request, --enabled, or --disabled")
	}
	if u.SetsTrigger() && u.AppName != "" {
		return fmt.Errorf("--app cannot be combined with --branch or --pull-request: an environment follows one branch or pull request for every app from a repo")
	}
	return nil
}

// SetsTrigger reports whether this update changes what the environment follows for the repo
func (u envAppsUpdate) SetsTrigger() bool {
	return u.BranchName != nil || u.PullRequestId != nil
}

// Select returns the apps this update applies to, ordered by name.
// Repo and app name comparison is case-insensitive because git hosts treat owner/name that way.
func (u envAppsUpdate) Select(apps []types.Application) []types.Application {
	matched := make([]types.Application, 0)
	for _, app := range apps {
		if !strings.EqualFold(app.Repo, u.Repo) {
			continue
		}
		if u.AppName != "" && !strings.EqualFold(app.Name, u.AppName) {
			continue
		}
		matched = append(matched, app)
	}
	sort.SliceStable(matched, func(i, j int) bool { return matched[i].Name < matched[j].Name })
	return matched
}

// Trigger is what the environment should follow for the repo of the matched apps
func (u envAppsUpdate) Trigger(matched []types.Application) (api.SetEnvGitopsTriggerInput, error) {
	repoUrl := ""
	for _, app := range matched {
		if app.RepoUrl != "" {
			repoUrl = app.RepoUrl
			break
		}
	}
	if repoUrl == "" {
		return api.SetEnvGitopsTriggerInput{}, fmt.Errorf("apps from repo %q have no repository url", u.Repo)
	}
	input := api.SetEnvGitopsTriggerInput{RepoUrl: repoUrl}
	if u.BranchName != nil {
		input.GitBranch = *u.BranchName
	}
	input.PullRequestId = u.PullRequestId
	return input, nil
}

// Automations returns the automation of each matched app with only the enabled flag changed.
// An app with no current automation is written with auto-deploy off.
func (u envAppsUpdate) Automations(matched []types.Application, current []types.WorkspaceAutomation) []api.UpdateWorkspaceAutomationInput {
	autoDeploy := map[int64]bool{}
	for _, automation := range current {
		autoDeploy[automation.BlockId] = automation.AutoDeploy
	}
	inputs := make([]api.UpdateWorkspaceAutomationInput, 0, len(matched))
	for _, app := range matched {
		inputs = append(inputs, api.UpdateWorkspaceAutomationInput{
			BlockId:    app.Id,
			IsEnabled:  u.Enable,
			AutoDeploy: autoDeploy[app.Id],
		})
	}
	return inputs
}

func (u envAppsUpdate) describeTrigger() string {
	if u.PullRequestId != nil {
		return fmt.Sprintf("pull request %d", *u.PullRequestId)
	}
	if u.BranchName != nil {
		return fmt.Sprintf("branch %s", *u.BranchName)
	}
	return ""
}

func describeTracking(previewApp types.PreviewApp) string {
	switch {
	case previewApp.BranchName != nil:
		return fmt.Sprintf("branch %s", *previewApp.BranchName)
	case previewApp.PullRequestId != nil:
		return fmt.Sprintf("pull request %d", *previewApp.PullRequestId)
	case previewApp.TriggerType != "":
		// a tag or a release: the env follows the repo without a branch or pull request
		return strings.ReplaceAll(previewApp.TriggerType, "_", " ")
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
