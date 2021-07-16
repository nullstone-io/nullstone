package app

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type StatusReport struct {
	Fields []string
	Data   map[string]interface{}
}

type StatusDetailReport struct {
	Fields []string
	Data   map[string]interface{}
}

// Provider provides a standard interface to run commands against an app
// Each Operator is responsible for:
//   - Collecting necessary information from a workspace's outputs
//   - Modifying infrastructure to perform each command (e.g. push, deploy, etc.)
//   - Each Provider is specific to Category+Type (Example: category=app/container, type=service/aws-fargate)
type Provider interface {
	DefaultLogProvider() string
	Push(nsConfig api.Config, details Details, userConfig map[string]string) error
	Deploy(nsConfig api.Config, details Details, userConfig map[string]string) error

	// Status returns a high-level status report on the specified app env
	Status(nsConfig api.Config, details Details) (StatusReport, error)

	// StatusDetail returns a detailed status report on the specified app env
	StatusDetail(nsConfig api.Config, details Details) (StatusDetailReport, error)
}
