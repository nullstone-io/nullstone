package aws_ecs_fargate

import (
	"context"
	"encoding/json"
	"fmt"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	aws_ecr "gopkg.in/nullstone-io/nullstone.v0/app/container/aws-ecr"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
	"time"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppContainer),
	Provider:    "aws",
	Platform:    "ecs",
	Subplatform: "fargate",
}

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "cloudwatch"
}

func (p Provider) identify(nsConfig api.Config, details app.Details) (*InfraConfig, error) {
	logger.Printf("Identifying infrastructure for app %q\n", details.App.Name)
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)
	return ic, nil
}

func (p Provider) Push(nsConfig api.Config, details app.Details, source, version string) error {
	return (aws_ecr.Provider{}).Push(nsConfig, details, source, version)
}

// Deploy takes the following steps to deploy an AWS ECS service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (p Provider) Deploy(nsConfig api.Config, details app.Details, version string) (*string, error) {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return nil, err
	}

	taskDef, err := ic.GetTaskDefinition()
	if err != nil {
		return nil, fmt.Errorf("error retrieving current service information: %w", err)
	}

	logger.Printf("Deploying app %q\n", details.App.Name)
	if version == "" {
		return nil, fmt.Errorf("no version specified, version is required to deploy")
	}
	taskDefArn := *taskDef.TaskDefinitionArn
	logger.Printf("Updating app version to %q\n", version)
	logger.Printf("Updating image tag to %q\n", version)
	newTaskDef, err := ic.UpdateTaskImageTag(taskDef, version)
	if err != nil {
		return nil, fmt.Errorf("error updating task with new image tag: %w", err)
	}
	taskDefArn = *newTaskDef.TaskDefinitionArn

	deployment, err := ic.UpdateServiceTask(taskDefArn)
	if err != nil {
		return nil, fmt.Errorf("error deploying service: %w", err)
	}

	log.Println(fmt.Sprintf("Associating deploy with deployment: %s", *deployment.Id))
	logger.Printf("Deployed app %q\n", details.App.Name)
	return deployment.Id, nil
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

	report := app.StatusReport{
		Fields: []string{"Id", "Running", "Desired", "Pending"},
		Data: map[string]interface{}{
			"Running": fmt.Sprintf("%d", svc.RunningCount),
			"Desired": fmt.Sprintf("%d", svc.DesiredCount),
			"Pending": fmt.Sprintf("%d", svc.PendingCount),
		},
	}
	return report, nil
}

func (p Provider) DeploymentStatus(deploymentId string, nsConfig api.Config, details app.Details) (app.RolloutStatus, string, []string, error) {
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(details.Workspace, &ic.Outputs); err != nil {
		return app.RolloutStatusUnknown, "", nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}

	deployment, err := ic.GetDeployment(deploymentId)
	if err != nil {
		return app.RolloutStatusUnknown, "", nil, err
	}
	rolloutStatus, message := getRolloutStatus(deployment)
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(deployment)

	tasks, err := ic.GetDeploymentTasks(deploymentId)
	if err != nil {
		return app.RolloutStatusUnknown, "", nil, err
	}
	taskMessages := make([]string, len(tasks))
	for i, task := range tasks {
		encoder.Encode(task)
		taskMessages[i] = formatTaskStatus(task)
	}

	return rolloutStatus, message, taskMessages, nil
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
			Fields: []string{
				"Id",
				"Created",
				"Status",
				"Running",
				"Desired",
				"Pending",
				"Rollout Status",
				"Rollout Status Reason",
			},
			Data: map[string]interface{}{
				"Id":                    deployment.Id,
				"Created":               fmt.Sprintf("%s", *deployment.CreatedAt),
				"Status":                *deployment.Status,
				"Running":               fmt.Sprintf("%d", deployment.RunningCount),
				"Desired":               fmt.Sprintf("%d", deployment.DesiredCount),
				"Pending":               fmt.Sprintf("%d", deployment.PendingCount),
				"Rollout Status":        deployment.RolloutState,
				"Rollout Status Reason": deployment.RolloutStateReason,
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

func getRolloutStatus(deployment *ecstypes.Deployment) (app.RolloutStatus, string) {
	status := app.RolloutStatusUnknown
	if deployment.RolloutState == "IN_PROGRESS" {
		status = app.RolloutStatusInProgress
	} else if deployment.RolloutState == "COMPLETED" {
		status = app.RolloutStatusComplete
	} else if deployment.RolloutState == "FAILED" {
		status = app.RolloutStatusFailed
	}
	message := *deployment.RolloutStateReason
	if deployment.RunningCount == deployment.DesiredCount {
		message = fmt.Sprintf("%s All %d services are running.", message, deployment.RunningCount)
	} else if deployment.DesiredCount > 0 {
		message = fmt.Sprintf("%s %d running out of %d services are running.", message, deployment.RunningCount, deployment.DesiredCount)
	} else {
		message = fmt.Sprintf("%s Not attempting to start any services.", message)
	}
	return status, message
}

func formatTaskStatus(task ecstypes.Task) string {
	return fmt.Sprintf("%s - %s - %s - %s - %s - %s - %s - %s",
		task.HealthStatus,
		derefString(task.DesiredStatus),
		derefString(task.LastStatus),
		formatTime(task.PullStartedAt),
		formatTime(task.PullStoppedAt),
		formatTime(task.StoppingAt),
		formatTime(task.StoppedAt),
		derefString(task.StoppedReason),
	)
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func formatTime(t *time.Time) string {
	if t != nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return ""
}
