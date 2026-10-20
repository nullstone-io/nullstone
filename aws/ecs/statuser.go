package ecs

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	sdkecs "github.com/nullstone-io/deployment-sdk/aws/ecs"
	"github.com/nullstone-io/deployment-sdk/display"
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

// isJob reports whether this workspace is a one-off task app (no ECS service to describe).
func (s Statuser) isJob() bool {
	return s.Infra.ServiceName == ""
}

func (s Statuser) Status(ctx context.Context) (admin.StatusReport, error) {
	if s.isJob() {
		return s.jobStatus(ctx)
	}

	svc, err := GetService(ctx, s.Infra)
	if err != nil {
		return admin.StatusReport{}, fmt.Errorf("error retrieving ecs service: %w", err)
	}
	if svc == nil {
		return admin.StatusReport{}, fmt.Errorf("ecs service %q was not found", s.Infra.ServiceName)
	}

	return admin.StatusReport{
		Fields: []string{"Running", "Desired", "Pending"},
		Data: map[string]interface{}{
			"Running": fmt.Sprintf("%d", svc.RunningCount),
			"Desired": fmt.Sprintf("%d", svc.DesiredCount),
			"Pending": fmt.Sprintf("%d", svc.PendingCount),
		},
	}, nil
}

func (s Statuser) StatusDetail(ctx context.Context) (admin.StatusDetailReports, error) {
	if s.isJob() {
		return s.jobStatusDetail(ctx)
	}

	reports := admin.StatusDetailReports{}

	svc, err := GetService(ctx, s.Infra)
	if err != nil {
		return reports, fmt.Errorf("error retrieving ecs service: %w", err)
	}
	if svc == nil {
		return reports, fmt.Errorf("ecs service %q was not found", s.Infra.ServiceName)
	}

	deploymentReport := admin.StatusDetailReport{
		Name:    "Deployments",
		Records: admin.StatusRecords{},
	}
	for _, deployment := range svc.Deployments {
		record := admin.StatusRecord{
			Fields: []string{"Created", "Status", "Running", "Desired", "Pending"},
			Data: map[string]interface{}{
				"Created": display.FormatTimePtr(deployment.CreatedAt),
				"Status":  *deployment.Status,
				"Running": fmt.Sprintf("%d", deployment.RunningCount),
				"Desired": fmt.Sprintf("%d", deployment.DesiredCount),
				"Pending": fmt.Sprintf("%d", deployment.PendingCount),
			},
		}
		deploymentReport.Records = append(deploymentReport.Records, record)
	}
	reports = append(reports, deploymentReport)

	lbReport := admin.StatusDetailReport{
		Name:    "Load Balancers",
		Records: admin.StatusRecords{},
	}
	for _, lb := range svc.LoadBalancers {
		targets, err := GetTargetGroupHealth(ctx, s.Infra, *lb.TargetGroupArn)
		if err != nil {
			return reports, fmt.Errorf("error retrieving load balancer target health: %w", err)
		}

		for _, target := range targets {
			record := admin.StatusRecord{
				Fields: []string{"Port", "Target", "Status"},
				Data:   map[string]interface{}{"Port": fmt.Sprintf("%d", *lb.ContainerPort)},
			}
			record.Data["Target"] = *target.Target.Id
			record.Data["Status"] = target.TargetHealth.State
			if target.TargetHealth.Reason != "" {
				record.Fields = append(record.Fields, "Reason")
				record.Data["Reason"] = target.TargetHealth.Reason
			}

			lbReport.Records = append(lbReport.Records, record)
		}
	}
	reports = append(reports, lbReport)

	return reports, nil
}

// jobExecutions reuses the deployment-sdk status path for task apps, which merges live tasks with
// the cluster's task execution history (when the cluster module records it).
func (s Statuser) jobExecutions(ctx context.Context) ([]sdkecs.EcsJobExecution, error) {
	statuser := sdkecs.Statuser{
		OsWriters: s.OsWriters,
		Details:   s.Details,
		Infra:     s.Infra.SdkOutputs(),
	}
	result, err := statuser.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving ecs task executions: %w", err)
	}
	status, ok := result.(sdkecs.Status)
	if !ok {
		return nil, fmt.Errorf("unexpected ecs status result %T", result)
	}
	return status.Executions, nil
}

func (s Statuser) jobStatus(ctx context.Context) (admin.StatusReport, error) {
	executions, err := s.jobExecutions(ctx)
	if err != nil {
		return admin.StatusReport{}, err
	}

	running := 0
	for _, exec := range executions {
		if exec.Phase == sdkecs.EcsJobPhaseRunning || exec.Phase == sdkecs.EcsJobPhaseQueued {
			running++
		}
	}
	report := admin.StatusReport{
		Fields: []string{"Running", "Executions", "Last Run", "Last Result"},
		Data: map[string]interface{}{
			"Running":     fmt.Sprintf("%d", running),
			"Executions":  fmt.Sprintf("%d", len(executions)),
			"Last Run":    "",
			"Last Result": "",
		},
	}
	// Executions arrive running-first, then newest stopped first.
	if len(executions) > 0 {
		latest := executions[0]
		report.Data["Last Run"] = display.FormatTimePtr(latest.StartedAt)
		report.Data["Last Result"] = string(latest.Phase)
	}
	return report, nil
}

func (s Statuser) jobStatusDetail(ctx context.Context) (admin.StatusDetailReports, error) {
	executions, err := s.jobExecutions(ctx)
	if err != nil {
		return nil, err
	}

	report := admin.StatusDetailReport{
		Name:    "Executions",
		Records: admin.StatusRecords{},
	}
	for _, exec := range executions {
		record := admin.StatusRecord{
			Fields: []string{"Started", "Stopped", "Status", "Version", "Exit Code"},
			Data: map[string]interface{}{
				"Started":   display.FormatTimePtr(exec.StartedAt),
				"Stopped":   display.FormatTimePtr(exec.StoppedAt),
				"Status":    string(exec.Phase),
				"Version":   exec.AppVersion,
				"Exit Code": mainExitCode(exec, s.Infra.MainContainerName),
			},
		}
		if exec.Phase == sdkecs.EcsJobPhaseFailed && exec.StoppedReason != "" {
			record.Fields = append(record.Fields, "Reason")
			record.Data["Reason"] = exec.StoppedReason
		}
		report.Records = append(report.Records, record)
	}
	return admin.StatusDetailReports{report}, nil
}

// mainExitCode is the main container's exit code, or the first container's when the main
// container is unknown. Empty until the container has exited.
func mainExitCode(exec sdkecs.EcsJobExecution, mainContainerName string) string {
	for _, c := range exec.Containers {
		if mainContainerName != "" && c.Name != mainContainerName {
			continue
		}
		if c.ExitCode != nil {
			return fmt.Sprintf("%d", *c.ExitCode)
		}
		return ""
	}
	return ""
}
