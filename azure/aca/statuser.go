package aca

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

func NewStatuser(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Statuser, error) {
	outs, err := outputs.Retrieve[Outputs](ctx, source, appDetails.Workspace, appDetails.WorkspaceConfig)
	if err != nil {
		return nil, err
	}
	outs.InitializeCreds(source, appDetails.Workspace)

	return Statuser{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type Statuser struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     Outputs
}

func (s Statuser) Status(ctx context.Context) (admin.StatusReport, error) {
	if s.Infra.ContainerAppName == "" {
		return admin.StatusReport{}, fmt.Errorf("container app name is not configured")
	}

	replicas, err := GetReplicas(ctx, s.Infra)
	if err != nil {
		return admin.StatusReport{}, fmt.Errorf("error retrieving container app replicas: %w", err)
	}

	running := 0
	for _, r := range replicas {
		if r.Running {
			running++
		}
	}

	return admin.StatusReport{
		Fields: []string{"Running", "Total"},
		Data: map[string]interface{}{
			"Running": fmt.Sprintf("%d", running),
			"Total":   fmt.Sprintf("%d", len(replicas)),
		},
	}, nil
}

func (s Statuser) StatusDetail(ctx context.Context) (admin.StatusDetailReports, error) {
	if s.Infra.ContainerAppName == "" {
		return nil, fmt.Errorf("container app name is not configured")
	}

	replicas, err := GetReplicas(ctx, s.Infra)
	if err != nil {
		return nil, fmt.Errorf("error retrieving container app replicas: %w", err)
	}

	replicaReport := admin.StatusDetailReport{
		Name:    "Replicas",
		Records: admin.StatusRecords{},
	}
	for _, r := range replicas {
		status := "Not Running"
		if r.Running {
			status = "Running"
		}
		record := admin.StatusRecord{
			Fields: []string{"Name", "Status", "Created"},
			Data: map[string]interface{}{
				"Name":    r.Name,
				"Status":  status,
				"Created": r.CreatedAt,
			},
		}
		replicaReport.Records = append(replicaReport.Records, record)
	}

	return admin.StatusDetailReports{replicaReport}, nil
}
