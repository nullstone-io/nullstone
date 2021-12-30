package gcp_gke

import (
	"gopkg.in/nullstone-io/nullstone.v0/gcp"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
)

type Outputs struct {
	ClusterId            string             `ns:"cluster_id"`
	ClusterEndpoint      string             `ns:"cluster_endpoint"`
	ClusterCACertificate string             `ns:"cluster_ca_certificate"`
	Deployer             gcp.ServiceAccount `ns:"deployer"`
}

func (o Outputs) ClusterInfo() k8s.ClusterInfo {
	return k8s.ClusterInfo{
		ID:            o.ClusterId,
		Endpoint:      o.ClusterEndpoint,
		CACertificate: o.ClusterCACertificate,
	}
}
