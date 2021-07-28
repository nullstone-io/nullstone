package gcp_gke_service

import gcp_gke "gopkg.in/nullstone-io/nullstone.v0/contracts/gcp-gke"

type Outputs struct {
	Namespace         string `ns:"namespace"`
	Name              string `ns:"name"`
	MainContainerName string `ns:"main_container_name,optional"`

	Cluster gcp_gke.Outputs `ns:",connectionType:cluster/aws-fargate"`
}
