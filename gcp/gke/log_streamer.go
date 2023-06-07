package gke

import (
	"context"
	"github.com/fatih/color"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/display"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	"io"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger     = log.New(os.Stderr, "", 0)
	infoLogger = log.New(logger.Writer(), "    ", 0)
	bold       = color.New(color.Bold)
	normal     = color.New()
)

func NewLogStreamer(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.LogStreamer, error) {
	outs, err := outputs.Retrieve[Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	emitter := func(w io.Writer, podName, containerName, line string) error {
		// Try to read the timestamp at the front of the line
		// If found, parse it to format in the user's timezone
		timestamp, remaining := cutTimestampPrefix(line)
		if timestamp != nil {
			normal.Fprint(w, display.FormatTimePtr(timestamp))
			normal.Fprint(w, " ")
		}
		bold.Fprintf(w, "[%s/%s] ", podName, containerName)
		normal.Fprint(w, remaining)
		return nil
	}

	return k8s.LogStreamer{
		OsWriters:    osWriters,
		Details:      appDetails,
		AppNamespace: outs.ServiceNamespace,
		AppName:      outs.ServiceName,
		NewConfigFn: func(ctx context.Context) (*rest.Config, error) {
			return gke.CreateKubeConfig(ctx, outs.ClusterNamespace, outs.Deployer)
		},
		Emitter: emitter,
	}, nil
}

func cutTimestampPrefix(line string) (*time.Time, string) {
	if before, after, found := strings.Cut(line, ""); found {
		if ts, parseErr := time.Parse(time.RFC3339, before); parseErr == nil {
			return &ts, after
		}
	}
	return nil, line
}
