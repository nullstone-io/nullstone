package app

import "gopkg.in/nullstone-io/go-api-client.v0/types"

type Details struct {
	App       *types.Application
	Env       *types.Environment
	Workspace *types.Workspace
}
