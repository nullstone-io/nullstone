package ecs

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/creds"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	Region             string     `ns:"region"`
	ServiceName        string     `ns:"service_name"`
	TaskArn            string     `ns:"task_arn"`
	MainContainerName  string     `ns:"main_container_name,optional"`
	Deployer           nsaws.User `ns:"deployer,optional"`
	AppSecurityGroupId string     `ns:"app_security_group_id"`
	LaunchType         string     `ns:"launch_type,optional"`

	Cluster          ClusterOutputs          `ns:",connectionContract:cluster/aws/ecs:*,optional"`
	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/aws/ecs:*,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	credsFactory := creds.NewProviderFactory(source, ws.StackId, ws.Uid)
	o.Deployer.RemoteProvider = credsFactory("adminer", "deployer")
}

func (o *Outputs) ClusterArn() string {
	if o.ClusterNamespace.ClusterArn != "" {
		return o.ClusterNamespace.ClusterArn
	}
	return o.Cluster.ClusterArn
}

func (o *Outputs) PrivateSubnetIds() []string {
	if o.ClusterNamespace.Cluster.Network.PrivateSubnetIds != nil {
		return o.ClusterNamespace.Cluster.Network.PrivateSubnetIds
	}
	return o.Cluster.Network.PrivateSubnetIds
}

func (o *Outputs) GetLaunchType() types.LaunchType {
	switch o.LaunchType {
	case "EC2":
		return types.LaunchTypeEc2
	case "EXTERNAL":
		return types.LaunchTypeExternal
	default:
		return types.LaunchTypeFargate
	}
}

type ClusterNamespaceOutputs struct {
	ClusterArn string         `ns:"cluster_arn"`
	Cluster    ClusterOutputs `ns:"connectionContract:cluster/aws/ecs:*,optional"`
}

type ClusterOutputs struct {
	ClusterArn string         `ns:"cluster_arn"`
	Network    NetworkOutputs `ns:"connectionContract:network/aws/*,optional"`
}

type NetworkOutputs struct {
	PrivateSubnetIds []string `ns:"private_subnet_ids"`
}
