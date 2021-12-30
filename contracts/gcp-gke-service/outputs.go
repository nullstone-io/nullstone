package gcp_gke_service

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	gcp_gke "gopkg.in/nullstone-io/nullstone.v0/contracts/gcp-gke"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

type Outputs struct {
	ServiceNamespace  string          `ns:"service_namespace"`
	ServiceName       string          `ns:"service_name"`
	ImageRepoUrl      docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher       aws.User        `ns:"image_pusher,optional"`
	MainContainerName string          `ns:"main_container_name,optional"`

	Cluster gcp_gke.Outputs `ns:",connectionType:cluster/gcp-gke"`
}
