package gcp_gke

import "gopkg.in/nullstone-io/nullstone.v0/gcp"

type Outputs struct {
	ClusterId            string             `ns:"cluster_id"`
	ClusterEndpoint      string             `ns:"cluster_endpoint"`
	ClusterCACertificate string             `ns:"cluster_ca_certificate"`
	Deployer             gcp.ServiceAccount `ns:"deployer"`
}
