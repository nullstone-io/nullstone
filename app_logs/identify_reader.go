package app_logs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

type outputsLogProvider struct {
	LogProvider string `ns:"log_provider,optional"`
}

func IdentifyReader(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (string, error) {
	logger.Printf("Identifying infrastructure for app %q\n", app.Name)
	lpOutputs := &outputsLogProvider{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(workspace, &lpOutputs); err != nil {
		return lpOutputs.LogProvider, fmt.Errorf("Unable to identify app logger: %w", err)
	}
	return "", nil
}
