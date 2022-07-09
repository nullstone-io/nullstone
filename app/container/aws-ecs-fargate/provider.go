package aws_ecs_fargate

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecr"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecs"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
)

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppContainer),
	Provider:    "aws",
	Platform:    "ecs",
	Subplatform: "fargate",
}

func NewProvider(logger *log.Logger, nsConfig api.Config, appDetails app.Details) app.Provider {
	return Provider{
		Logger:     logger,
		NsConfig:   nsConfig,
		AppDetails: appDetails,
	}
}

type Provider struct {
	Logger     *log.Logger
	NsConfig   api.Config
	AppDetails app.Details
}

func (p Provider) NewPusher() (app.Pusher, error) {
	return ecr.NewPusher(p.Logger, p.NsConfig, p.AppDetails)
}

func (p Provider) NewDeployer() (app.Deployer, error) {
	return ecs.NewDeployer(p.Logger, p.NsConfig, p.AppDetails)
}

func (p Provider) NewDeployStatusGetter() (app.DeployStatusGetter, error) {
	//TODO implement me
	panic("implement me")
}

func (p Provider) DefaultLogProvider() string {
	return "cloudwatch"
}

func (p Provider) Exec(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	task := userConfig["task"]
	if task == "" {
		if task, err = ic.GetRandomTask(); err != nil {
			return err
		} else if task == "" {
			return fmt.Errorf("cannot exec command with no running tasks")
		}
	}

	return ic.ExecCommand(ctx, task, userConfig["cmd"], nil)
}

func (p Provider) Ssh(ctx context.Context, nsConfig api.Config, details app.Details, userConfig map[string]any) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	task, _ := userConfig["task"].(string)
	if task == "" {
		if task, err = ic.GetRandomTask(); err != nil {
			return err
		} else if task == "" {
			return fmt.Errorf("cannot exec command with no running tasks")
		}
	}

	if forwards, ok := userConfig["forwards"].([]config.PortForward); ok && len(forwards) > 0 {
		return fmt.Errorf("ecs:fargate provider does not support port forwarding")
	}

	return ic.ExecCommand(ctx, task, "/bin/sh", nil)
}

func (p Provider) Status(nsConfig api.Config, details app.Details) (app.StatusReport, error) {
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return app.StatusReport{}, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}

	svc, err := ic.GetService()
	if err != nil {
		return app.StatusReport{}, fmt.Errorf("error retrieving fargate service: %w", err)
	}

	return app.StatusReport{
		Fields: []string{"Running", "Desired", "Pending"},
		Data: map[string]interface{}{
			"Running": fmt.Sprintf("%d", svc.RunningCount),
			"Desired": fmt.Sprintf("%d", svc.DesiredCount),
			"Pending": fmt.Sprintf("%d", svc.PendingCount),
		},
	}, nil
}

func (p Provider) StatusDetail(nsConfig api.Config, details app.Details) (app.StatusDetailReports, error) {
	reports := app.StatusDetailReports{}

	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return reports, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}

	svc, err := ic.GetService()
	if err != nil {
		return reports, fmt.Errorf("error retrieving fargate service: %w", err)
	}

	deploymentReport := app.StatusDetailReport{
		Name:    "Deployments",
		Records: app.StatusRecords{},
	}
	for _, deployment := range svc.Deployments {
		record := app.StatusRecord{
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

	lbReport := app.StatusDetailReport{
		Name:    "Load Balancers",
		Records: app.StatusRecords{},
	}
	for _, lb := range svc.LoadBalancers {
		targets, err := ic.GetTargetGroupHealth(*lb.TargetGroupArn)
		if err != nil {
			return reports, fmt.Errorf("error retrieving load balancer target health: %w", err)
		}

		for _, target := range targets {
			record := app.StatusRecord{
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
