package ecs

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

func NewStatuser(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.Statuser, error) {
	outs, err := outputs.Retrieve[Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

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
	svc, err := GetService(ctx, s.Infra)
	if err != nil {
		return admin.StatusReport{}, fmt.Errorf("error retrieving ecs service: %w", err)
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
	reports := admin.StatusDetailReports{}

	svc, err := GetService(ctx, s.Infra)
	if err != nil {
		return reports, fmt.Errorf("error retrieving ecs service: %w", err)
	}

	deploymentReport := admin.StatusDetailReport{
		Name:    "Deployments",
		Records: admin.StatusRecords{},
	}
	for _, deployment := range svc.Deployments {
		record := admin.StatusRecord{
			Fields: []string{"Created", "Status", "Running", "Desired", "Pending"},
			Data: map[string]interface{}{
				"Created": fmt.Sprintf("%s", *deployment.CreatedAt),
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
