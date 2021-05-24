package app

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// Provider provides a standard interface to run commands against an app
// Each Operator is responsible for:
//   - Collecting necessary information from a workspace's outputs
//   - Modifying infrastructure to perform each command (e.g. push, deploy, etc.)
//   - Each Provider is specific to Category+Type (Example: category=app/container, type=service/aws-fargate)
type Provider interface {
	Push(nsConfig api.Config, app *types.Application, env *types.Environment, workspace *types.Workspace, userConfig map[string]string) error
	Deploy(nsConfig api.Config, app *types.Application, env *types.Environment, workspace *types.Workspace, userConfig map[string]string) error
}
