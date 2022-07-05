package app

import (
	"context"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"time"
)

type RolloutStatus string

const (
	RolloutStatusComplete   RolloutStatus = "complete"
	RolloutStatusInProgress               = "in-progress"
	RolloutStatusFailed                   = "failed"
	RolloutStatusUnknown                  = "unknown"
)

type ServiceEvent struct {
	Id        string
	CreatedAt time.Time
	Message   string
}

type StatusReport struct {
	Status  RolloutStatus
	Message string
	Fields  []string
	Data    map[string]interface{}
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
	Push(nsConfig api.Config, details Details, userConfig map[string]string) error
	Deploy(nsConfig api.Config, details Details, userConfig map[string]string) error

	// Exec allows a user to execute a command (usually tunneling) into a running service
	// This only makes sense for container-based providers
	Exec(ctx context.Context, nsConfig api.Config, details Details, userConfig map[string]string) error

	// Ssh allows a user to SSH into a running service
	Ssh(ctx context.Context, nsConfig api.Config, details Details, userConfig map[string]any) error

	// Status returns a high-level status report on the specified app env
	Status(nsConfig api.Config, details Details) (StatusReport, error)

	// DeploymentStatus returns the status of a specific deployment, the other status methods are summaries
	DeploymentStatus(deployReference string, nsConfig api.Config, details Details) (StatusReport, []ServiceEvent, error)

	// StatusDetail returns a detailed status report on the specified app env
	StatusDetail(nsConfig api.Config, details Details) (StatusDetailReports, error)
}
