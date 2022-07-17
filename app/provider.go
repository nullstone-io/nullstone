package app

import (
	"context"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
)

type RolloutStatus string

const (
	RolloutStatusComplete   RolloutStatus = "complete"
	RolloutStatusInProgress RolloutStatus = "in-progress"
	RolloutStatusFailed     RolloutStatus = "failed"
	RolloutStatusUnknown    RolloutStatus = "unknown"
)

type StatusReport struct {
	Fields []string
	Data   map[string]interface{}
}

type StatusDetailReports []StatusDetailReport

type StatusDetailReport struct {
	Name    string
	Records StatusRecords
}

type StatusRecords []StatusRecord

type StatusRecord struct {
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
	Push(logger *log.Logger, nsConfig api.Config, details Details, source, version string) error
	Deploy(logger *log.Logger, nsConfig api.Config, details Details, version string) (*string, error)

	// Exec allows a user to execute a command (usually tunneling) into a running service
	// This only makes sense for container-based providers
	Exec(ctx context.Context, logger *log.Logger, nsConfig api.Config, details Details, userConfig map[string]string) error

	// Ssh allows a user to SSH into a running service
	Ssh(ctx context.Context, logger *log.Logger, nsConfig api.Config, details Details, userConfig map[string]any) error

	// Status returns a high-level status report on the specified app env
	Status(logger *log.Logger, nsConfig api.Config, details Details) (StatusReport, error)

	// DeploymentStatus returns the status of a specific deployment, the other status methods are summaries
	DeploymentStatus(logger *log.Logger, nsConfig api.Config, deployReference string, details Details) (RolloutStatus, error)

	// StatusDetail returns a detailed status report on the specified app env
	StatusDetail(logger *log.Logger, nsConfig api.Config, details Details) (StatusDetailReports, error)
}
