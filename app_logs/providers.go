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

type Providers map[string]Provider

type outputsLogProvider struct {
	LogProvider string `ns:"log_provider,optional"`
}

func (p Providers) Identify(defaultProvider string, nsConfig api.Config, app *types.Application, workspace *types.Workspace) (Provider, error) {
	logger.Printf("Identifying log provider for app %q\n", app.Name)
	lpOutputs := outputsLogProvider{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(workspace, &lpOutputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app logger: %w", err)
	}
	if lpOutputs.LogProvider == "" {
		lpOutputs.LogProvider = defaultProvider
	}

	provider := p.Find(lpOutputs.LogProvider)
	if provider == nil {
		return nil, fmt.Errorf("Unable to stream logs, this CLI does not support log provider %s", lpOutputs.LogProvider)
	}
	return provider, nil
}

func (p Providers) Find(providerName string) Provider {
	if logProvider, ok := p[providerName]; ok {
		return logProvider
	}
	return nil
}
