package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var PreviewApps = &cli.Command{
	Name:      "preview-apps",
	Usage:     "View preview applications",
	UsageText: "nullstone preview-apps [subcommand]",
	Subcommands: []*cli.Command{
		PreviewAppsFind,
	},
}

// PreviewAppFound decorates a types.PreviewApp with the stack, env, and app names.
// The API returns ids only; the names are what the other CLI commands accept.
type PreviewAppFound struct {
	Stack             string `json:"stack"`
	StackId           int64  `json:"stackId"`
	Env               string `json:"env"`
	EnvId             int64  `json:"envId"`
	App               string `json:"app"`
	AppId             int64  `json:"appId"`
	Repo              string `json:"repo"`
	PullRequestId     *int64 `json:"pullRequestId"`
	PullRequestNumber *int   `json:"pullRequestNumber"`
}

var PreviewAppsFind = &cli.Command{
	Name: "find",
	Description: `Searches a stack for preview apps matching a repo and pull request.
This is used in CI to discover the preview environment created for a pull request without knowing the environment name up front.
The results report the stack, environment, and application names, which can be fed into other commands such as ` + "`nullstone wait`" + ` and ` + "`nullstone run`" + `.
If no preview apps match, this command writes a message to stderr and exits with a non-zero status so CI can branch on the result.`,
	Usage:     "Find preview apps by repo and pull request",
	UsageText: `nullstone preview-apps find --stack=<stack-name> --repo=<repo> [--pull-request=<pull-request>] [--format=table|json]`,
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.StringFlag{
			Name:     "repo",
			Usage:    "Filter preview apps by repository. Accepts either the repo name (e.g. acme/widgets) or the full repo URL.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "pull-request",
			Usage: "Filter preview apps by pull request. Accepts either the pull request number (e.g. 123) or the Nullstone pull request id.",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format. One of: table (default), json",
			Value: "table",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()

		format := strings.ToLower(c.String("format"))
		if format != "table" && format != "json" {
			return fmt.Errorf("invalid --format %q: must be table or json", format)
		}

		input := api.FindPreviewAppsInput{Repo: c.String("repo")}
		if c.IsSet("pull-request") {
			pullRequest, err := parsePullRequest(c.String("pull-request"))
			if err != nil {
				return err
			}
			input.PullRequest = pullRequest
		}

		stackName := c.String(StackRequiredFlag.Name)

		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			previewApps, err := client.PreviewApps().FindByStackName(ctx, stackName, input)
			if err != nil {
				return fmt.Errorf("error searching preview apps: %w", err)
			}

			if len(previewApps) == 0 {
				return fmt.Errorf("no preview apps found in stack %q matching %s", stackName, describePreviewAppFilters(input))
			}

			found, err := decoratePreviewApps(ctx, client, stackName, previewApps)
			if err != nil {
				return err
			}

			if format == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(found)
			}
			writePreviewAppsTable(found)
			return nil
		})
	},
}

// parsePullRequest reads the --pull-request flag, which is either a pull request number or a
// Nullstone pull request id. The server matches against both, so we only validate that it is numeric.
func parsePullRequest(raw string) (*int64, error) {
	// Accept a leading '#' since that is how pull request numbers are commonly written
	normalized := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	pullRequest, err := strconv.ParseInt(normalized, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid --pull-request %q: must be a pull request number or id", raw)
	}
	return &pullRequest, nil
}

func describePreviewAppFilters(input api.FindPreviewAppsInput) string {
	if input.PullRequest == nil {
		return fmt.Sprintf("repo %q", input.Repo)
	}
	return fmt.Sprintf("repo %q and pull request %d", input.Repo, *input.PullRequest)
}

// decoratePreviewApps resolves the env and app names for each preview app.
// Every result comes from stackName, so only the env and app need looking up.
// Each lookup is cached because many preview apps typically share an env.
func decoratePreviewApps(ctx context.Context, client api.Client, stackName string, previewApps []types.PreviewApp) ([]PreviewAppFound, error) {
	envNames := map[int64]string{}
	appNames := map[int64]string{}

	found := make([]PreviewAppFound, 0, len(previewApps))
	for _, previewApp := range previewApps {
		if _, ok := envNames[previewApp.EnvId]; !ok {
			env, err := client.Environments().Get(ctx, previewApp.StackId, previewApp.EnvId, false)
			if err != nil {
				return nil, fmt.Errorf("error looking for environment %d: %w", previewApp.EnvId, err)
			} else if env == nil {
				return nil, fmt.Errorf("environment %d does not exist", previewApp.EnvId)
			}
			envNames[previewApp.EnvId] = env.Name
		}
		if _, ok := appNames[previewApp.AppId]; !ok {
			app, err := client.Apps().Get(ctx, previewApp.StackId, previewApp.AppId)
			if err != nil {
				return nil, fmt.Errorf("error looking for application %d: %w", previewApp.AppId, err)
			} else if app == nil {
				return nil, fmt.Errorf("application %d does not exist", previewApp.AppId)
			}
			appNames[previewApp.AppId] = app.Name
		}

		cur := PreviewAppFound{
			Stack:         stackName,
			StackId:       previewApp.StackId,
			Env:           envNames[previewApp.EnvId],
			EnvId:         previewApp.EnvId,
			App:           appNames[previewApp.AppId],
			AppId:         previewApp.AppId,
			Repo:          previewApp.Repo,
			PullRequestId: previewApp.PullRequestId,
		}
		if previewApp.PullRequest != nil {
			number := previewApp.PullRequest.Number
			cur.PullRequestNumber = &number
		}
		found = append(found, cur)
	}

	sort.SliceStable(found, func(i, j int) bool {
		if found[i].Env != found[j].Env {
			return found[i].Env < found[j].Env
		}
		return found[i].App < found[j].App
	})
	return found, nil
}

func writePreviewAppsTable(found []PreviewAppFound) {
	rows := make([]string, 0, len(found)+1)
	rows = append(rows, "stack|env|app|repo|pull-request")
	for _, cur := range found {
		pullRequest := "<none>"
		if cur.PullRequestNumber != nil {
			pullRequest = fmt.Sprintf("#%d", *cur.PullRequestNumber)
		} else if cur.PullRequestId != nil {
			pullRequest = strconv.FormatInt(*cur.PullRequestId, 10)
		}
		rows = append(rows, fmt.Sprintf("%s|%s|%s|%s|%s", cur.Stack, cur.Env, cur.App, cur.Repo, pullRequest))
	}
	fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
}
