package aws_fargate

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
	"strings"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

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

func (p Provider) Push(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	sourceUrl := docker.ParseImageUrl(userConfig["source"])

	targetUrl := ic.Outputs.ImageRepoUrl
	// NOTE: We expect --version from the user which is used as the image tag for the pushed image
	if imageTag := userConfig["version"]; imageTag != "" {
		targetUrl.Tag = imageTag
	} else {
		targetUrl.Tag = sourceUrl.Tag
	}
	if targetUrl.String() == "" {
		return fmt.Errorf("cannot push if 'image_repo_url' module output is missing")
	}
	if !strings.Contains(targetUrl.Registry, "ecr") &&
		!strings.Contains(targetUrl.Registry, "amazonaws.com") {
		return fmt.Errorf("this app only supports push to AWS ECR (image=%s)", targetUrl)
	}
	// NOTE: For now, we are assuming that the production docker image is hosted in ECR
	// This will likely need to be refactored to support pushing to other image registries
	if ic.Outputs.ImagePusher.AccessKeyId == "" {
		return fmt.Errorf("cannot push without an authorized user, make sure 'image_pusher' output is not empty")
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	targetAuth, err := ic.GetEcrLoginAuth()
	if err != nil {
		return fmt.Errorf("error retrieving image registry credentials: %w", err)
	}

	logger.Printf("Retagging %s => %s\n", sourceUrl.String(), targetUrl.String())
	if err := ic.RetagImage(ctx, sourceUrl, targetUrl); err != nil {
		return fmt.Errorf("error retagging image: %w", err)
	}

	logger.Printf("Pushing %s\n", targetUrl.String())
	if err := ic.PushImage(ctx, targetUrl, targetAuth); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}

// Deploy takes the following steps to deploy an AWS Fargate service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	taskDef, err := ic.GetTaskDefinition()
	if err != nil {
		return fmt.Errorf("error retrieving current service information: %w", err)
	}

	logger.Printf("Deploying app %q\n", details.App.Name)
	version := userConfig["version"]
	taskDefArn := *taskDef.TaskDefinitionArn
	if version != "" {
		logger.Printf("Updating app version to %q\n", version)
		if err := app.UpdateVersion(nsConfig, details.App.Id, details.Env.Name, version); err != nil {
			return fmt.Errorf("error updating app version in nullstone: %w", err)
		}

		logger.Printf("Updating image tag to %q\n", version)
		newTaskDef, err := ic.UpdateTaskImageTag(taskDef, version)
		if err != nil {
			return fmt.Errorf("error updating task with new image tag: %w", err)
		}
		taskDefArn = *newTaskDef.TaskDefinitionArn
	}

	if err := ic.UpdateServiceTask(taskDefArn); err != nil {
		return fmt.Errorf("error deploying service: %w", err)
	}

	logger.Printf("Deployed app %q\n", details.App.Name)
	return nil
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
		Fields: []string{"running", "desired", "pending"},
		Data: map[string]interface{}{
			"running": fmt.Sprintf("%d", svc.RunningCount),
			"desired": fmt.Sprintf("%d", svc.DesiredCount),
			"pending": fmt.Sprintf("%d", svc.PendingCount),
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
