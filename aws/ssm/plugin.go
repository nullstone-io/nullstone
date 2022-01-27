package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type EcsSession struct {
	Cluster       string
	TaskId        string
	ContainerName string
}

func StartEcsSession(session *ecstypes.Session, region, cluster, task, containerName string) error {
	sessionJsonRaw, _ := json.Marshal(session)
	targetRaw, _ := json.Marshal(ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", cluster, task, containerName)),
	})

	er := ecs.NewDefaultEndpointResolver()
	endpoint, _ := er.ResolveEndpoint(region, ecs.EndpointResolverOptions{})

	args := []string{
		string(sessionJsonRaw),
		region,
		"StartSession",
		"", // empty profile name
		string(targetRaw),
		endpoint.URL,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-sigs:
				cancel()
				return
			}
		}
	}()
	defer close(sigs)

	process, err := getSessionManagerPluginPath()
	if err != nil {
		return fmt.Errorf("could not find AWS session-manager-plugin: %w", err)
	}

	cmd := exec.CommandContext(ctx, process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

const (
	sessionManagerBinary = "session-manager-plugin"
	sessionManagerUrl    = "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
)

// getSessionManagerPluginPath attempts to find "session-manager-plugin"
// If it's in PATH, will simply return binary name
// If not, will attempt OS-specific locations
func getSessionManagerPluginPath() (string, error) {
	if _, err := exec.LookPath(sessionManagerBinary); err == nil {
		return sessionManagerBinary, nil
	}
	if _, err := os.Stat(osSessionManagerPluginPath); err != nil {
		return "", fmt.Errorf("Could not find session-manager-plugin. Visit %q to install.", sessionManagerUrl)
	}
	return osSessionManagerPluginPath, nil
}
