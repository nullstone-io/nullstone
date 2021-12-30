package gcp_gke

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	gcp_gcr "gopkg.in/nullstone-io/nullstone.v0/app/container/gcp-gcr"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) DefaultLogProvider() string {
	return "gcp"
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
	if targetUrl.String() == "" {
		return fmt.Errorf("cannot push if 'image_repo_url' module output is missing")
	}
	targetUrl.Tag = userConfig["version"]

	return gcp_gcr.PushImage(sourceUrl, targetUrl, ic.Outputs.ImagePusher)
}

// Deploy takes the following steps to deploy a GCP GKE pod
//   Get pod
//   Change image tag
//   Update pod
func (p Provider) Deploy(nsConfig api.Config, details app.Details, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	deployment, err := ic.GetDeployment()
	if err != nil {
		return fmt.Errorf("error retrieving deployment: %w", err)
	}

	podSpec := deployment.Spec.Template.Spec

	logger.Printf("Deploying app %q\n", details.App.Name)
	version := userConfig["version"]
	if version != "" {
		logger.Printf("Updating app version to %q\n", version)
		if err := app.UpdateVersion(nsConfig, details.App.Id, details.Env.Name, version); err != nil {
			return fmt.Errorf("error updating app version in nullstone: %w", err)
		}

		logger.Printf("Updating image tag to %q\n", version)
		if podSpec, err = k8s.SetContainerImageTag(podSpec, ic.Outputs.MainContainerName, version); err != nil {
			return fmt.Errorf("error updating pod spec with new image tag: %w", err)
		}
	}

	deployment.Spec.Template.Spec = podSpec
	if _, err := ic.UpdateDeployment(deployment); err != nil {
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

	deployment, err := ic.GetDeployment()
	if err != nil {
		return app.StatusReport{}, fmt.Errorf("error retrieving deployment: %w", err)
	}

	return app.StatusReport{
		Fields: []string{"Available", "Ready", "Unavailable"},
		Data: map[string]interface{}{
			"Available":   fmt.Sprintf("%d", deployment.Status.AvailableReplicas),
			"Ready":       fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
			"Unavailable": fmt.Sprintf("%d", deployment.Status.UnavailableReplicas),
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

	deployment, err := ic.GetDeployment()
	if err != nil {
		return reports, fmt.Errorf("error retrieving deployment: %w", err)
	}

	deploymentReport := app.StatusDetailReport{
		Name:    "Deployments",
		Records: app.StatusRecords{},
	}
	record := app.StatusRecord{
		Fields: []string{"Created", "Available", "Ready", "Unavailable"},
		Data: map[string]interface{}{
			"Created":     fmt.Sprintf("%s", deployment.CreationTimestamp),
			"Available":   fmt.Sprintf("%d", deployment.Status.AvailableReplicas),
			"Ready":       fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
			"Unavailable": fmt.Sprintf("%d", deployment.Status.UnavailableReplicas),
		},
	}
	deploymentReport.Records = append(deploymentReport.Records, record)
	reports = append(reports, deploymentReport)

	services, err := ic.GetServices()

	lbReport := app.StatusDetailReport{
		Name:    "Load Balancers",
		Records: app.StatusRecords{},
	}
	for _, service := range services.Items {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			for _, portStatus := range ingress.Ports {
				record := app.StatusRecord{
					Fields: []string{"Port", "Target", "Status"},
					Data:   map[string]interface{}{"Port": fmt.Sprintf("%d/%s", portStatus.Port, portStatus.Protocol)},
				}
				target := ingress.Hostname
				if target == "" {
					target = ingress.IP
				}
				record.Data["Target"] = target

				status := ""
				if portStatus.Error != nil {
					status = *portStatus.Error
				}
				record.Data["Status"] = status

				lbReport.Records = append(lbReport.Records, record)
			}
		}
	}
	reports = append(reports, lbReport)

	return reports, nil
}
