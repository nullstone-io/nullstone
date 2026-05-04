package cmd

import (
	"context"
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
	"regexp"
	"time"
)

// These regexes anchor and exclude separators (`,`, `=`, whitespace, etc.) so a
// value can't smuggle extra label clauses or URL-breaking characters through.
// The K8s variants match the validation enigma applies on the API side.

// logsPodTemplateHashRegex matches values k8s generates for the pod-template-hash label.
var logsPodTemplateHashRegex = regexp.MustCompile(`^[a-z0-9]{1,63}$`)

// logsDNS1123LabelRegex matches a k8s DNS-1123 label, the format used for job names.
var logsDNS1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$`)

// logsDNS1123SubdomainRegex matches a k8s DNS-1123 subdomain, the format used
// for pod names. Length is bounded separately at 253.
var logsDNS1123SubdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// logsEcsTaskIdRegex matches an ECS task ID. Native Fargate task IDs are 32
// lowercase hex chars; this is a slightly looser alphanumeric form to tolerate
// EC2 launch types and any future ID-format change.
var logsEcsTaskIdRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,128}$`)

// logsEcsDeploymentIdRegex matches an ECS deployment ID. The native form is
// "ecs-svc/<numeric>" so the regex permits one optional slash-delimited segment.
// Total length is bounded separately.
var logsEcsDeploymentIdRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?$`)

var Logs = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "logs",
		Description: "Streams an application's logs to the console for the given environment. Use the start-time `-s` and end-time `-e` flags to only show logs for a given time period. Use the tail flag `-t` to stream the logs in real time.",
		Usage:       "Emit application logs",
		UsageText:   "nullstone logs [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			&cli.DurationFlag{
				Name:        "start-time",
				Aliases:     []string{"s"},
				DefaultText: "0s",
				Usage: `
       Emit log events that occur after the specified start-time. 
       This is a golang duration relative to the time the command is issued.
       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)
      `,
			},
			&cli.DurationFlag{
				Name:    "end-time",
				Aliases: []string{"e"},
				Usage: `
       Emit log events that occur before the specified end-time. 
       This is a golang duration relative to the time the command is issued.
       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)
      `,
			},
			&cli.DurationFlag{
				Name:        "interval",
				DefaultText: "1s",
				Usage: `Set --interval to a golang duration to control how often to pull new log events.
       This will do nothing unless --tail is set. The default is '1s' (1 second).
      `,
			},
			&cli.BoolFlag{
				Name:    "tail",
				Aliases: []string{"t"},
				Usage: `Set tail to watch log events and emit as they are reported.
       Use --interval to control how often to query log events.
       This is off by default. Unless this option is provided, this command will exit as soon as current log events are emitted.`,
			},
			&cli.StringFlag{
				Name:  "pod",
				Usage: "Restrict logs to a single pod by name (Kubernetes only).",
			},
			&cli.StringFlag{
				Name:  "job",
				Usage: "Restrict logs to a single Kubernetes job by name (adds `job-name=<value>` to the selector) or a single ECS task by ID.",
			},
			&cli.StringFlag{
				Name:  "pod-template-hash",
				Usage: "Restrict logs to a single ReplicaSet revision (adds `pod-template-hash=<value>` to the selector).",
			},
			&cli.StringFlag{
				Name:  "task",
				Usage: "Restrict logs to a single ECS task by ID.",
			},
			&cli.StringFlag{
				Name:  "deployment",
				Usage: "Restrict logs to a single ECS deployment by ID; logs from all tasks under the deployment are included.",
			},
		},
		Action: func(c *cli.Context) error {
			logStreamOptions := app.LogStreamOptions{
				WatchInterval: -1 * time.Second, // Disabled by default
				Emitter:       app.NewWriterLogEmitter(os.Stdout),
			}
			if hash := c.String("pod-template-hash"); hash != "" {
				if !logsPodTemplateHashRegex.MatchString(hash) {
					return cli.Exit("--pod-template-hash must be 1-63 lowercase alphanumeric characters", 1)
				}
				logStreamOptions.Selectors = append(logStreamOptions.Selectors, fmt.Sprintf("pod-template-hash=%s", hash))
			}
			if jobName := c.String("job"); jobName != "" {
				// Accept either a K8s DNS-1123 label or an ECS task ID. The K8s task ID
				// format (32 hex chars) already passes the DNS-1123 regex, so check that
				// first and fall back to the looser ECS form for non-hex EC2 task IDs.
				if !logsDNS1123LabelRegex.MatchString(jobName) && !logsEcsTaskIdRegex.MatchString(jobName) {
					return cli.Exit("--job must be a valid Kubernetes job name (DNS-1123 label) or an ECS task ID (alphanumeric)", 1)
				}
				// K8s consumes this via Selectors; ECS reads it off LogStreamOptions.Job.
				// Each streamer ignores the irrelevant field.
				logStreamOptions.Selectors = append(logStreamOptions.Selectors, fmt.Sprintf("job-name=%s", jobName))
				logStreamOptions.Job = jobName
			}
			if pod := c.String("pod"); pod != "" {
				if len(pod) > 253 || !logsDNS1123SubdomainRegex.MatchString(pod) {
					return cli.Exit("--pod must be a valid DNS-1123 subdomain pod name (1-253 chars, lowercase alphanumeric, '-', '.')", 1)
				}
				logStreamOptions.Pod = pod
			}
			if task := c.String("task"); task != "" {
				if !logsEcsTaskIdRegex.MatchString(task) {
					return cli.Exit("--task must be a valid ECS task ID (1-128 alphanumeric characters)", 1)
				}
				logStreamOptions.Task = task
			}
			if deployment := c.String("deployment"); deployment != "" {
				if len(deployment) > 256 || !logsEcsDeploymentIdRegex.MatchString(deployment) {
					return cli.Exit("--deployment must be a valid ECS deployment ID (max 256 chars, alphanumeric plus '/', '-', '_')", 1)
				}
				logStreamOptions.Deployment = deployment
			}
			if c.IsSet("start-time") {
				absoluteTime := time.Now().Add(-c.Duration("start-time"))
				logStreamOptions.StartTime = &absoluteTime
			} else {
				absoluteTime := time.Now()
				logStreamOptions.StartTime = &absoluteTime
			}
			if c.IsSet("end-time") {
				absoluteTime := time.Now().Add(-c.Duration("end-time"))
				logStreamOptions.EndTime = &absoluteTime
			}
			if c.IsSet("tail") {
				logStreamOptions.WatchInterval = time.Duration(0)
				if c.IsSet("interval") {
					logStreamOptions.WatchInterval = c.Duration("interval")
				}
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				source := outputs.ApiRetrieverSource{Config: cfg}
				logStreamer, err := providers.FindLogStreamer(ctx, osWriters, source, appDetails)
				if err != nil {
					return err
				}
				if logStreamer == nil {
					colorstring.Fprintln(osWriters.Stderr(), "[yellow]log streaming is not supported for this provider[reset]")
				}
				return logStreamer.Stream(ctx, logStreamOptions)
			})
		},
	}
}
