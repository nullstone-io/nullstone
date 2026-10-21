package ecs

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/creds"
	sdkecs "github.com/nullstone-io/deployment-sdk/aws/ecs"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	Region             string            `ns:"region"`
	ServiceName        string            `ns:"service_name"`
	TaskArn            string            `ns:"task_arn"`
	MainContainerName  string            `ns:"main_container_name,optional"`
	Runner             nsaws.IamIdentity `ns:"deployer,optional"`
	Remoter            nsaws.IamIdentity `ns:"deployer,optional"`
	Statuser           nsaws.IamIdentity `ns:"deployer,optional"`
	AppSecurityGroupId string            `ns:"app_security_group_id"`
	LaunchType         string            `ns:"launch_type,optional"`

	Cluster          ClusterOutputs          `ns:",connectionContract:cluster/aws/ecs:*,optional"`
	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/aws/ecs:*,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	credsFactory := creds.NewProviderFactory(source, ws.StackId, ws.BlockId, ws.EnvId)
	o.Runner.RemoteProvider = credsFactory("adminer", "deployer")
	o.Remoter.RemoteProvider = credsFactory("adminer", "deployer")
	o.Statuser.RemoteProvider = credsFactory("adminer", "deployer")
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
	ClusterArn             string         `ns:"cluster_arn"`
	ExecutionsLogGroupName string         `ns:"executions_log_group_name,optional"`
	Cluster                ClusterOutputs `ns:"connectionContract:cluster/aws/ecs:*,optional"`
}

type ClusterOutputs struct {
	ClusterArn             string         `ns:"cluster_arn"`
	ExecutionsLogGroupName string         `ns:"executions_log_group_name,optional"`
	Network                NetworkOutputs `ns:"connectionContract:network/aws/*,optional"`
}

type NetworkOutputs struct {
	PrivateSubnetIds []string `ns:"private_subnet_ids"`
}

// SdkOutputs maps these outputs onto the deployment-sdk's ECS outputs so the CLI can reuse its
// status logic. The SDK's deployer identity is this CLI's statuser identity, so credentials keep
// the "adminer" purpose the CLI already uses for status calls.
func (o *Outputs) SdkOutputs() sdkecs.Outputs {
	return sdkecs.Outputs{
		Region:            o.Region,
		ServiceName:       o.ServiceName,
		TaskArn:           o.TaskArn,
		MainContainerName: o.MainContainerName,
		Deployer:          o.Statuser,
		Cluster: sdkecs.ClusterOutputs{
			ClusterArn:             o.Cluster.ClusterArn,
			ExecutionsLogGroupName: o.Cluster.ExecutionsLogGroupName,
		},
		ClusterNamespace: sdkecs.ClusterNamespaceOutputs{
			ClusterArn:             o.ClusterNamespace.ClusterArn,
			ExecutionsLogGroupName: o.ClusterNamespace.ExecutionsLogGroupName,
		},
	}
}
