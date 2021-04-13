package s3site

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/deploy"
	"gopkg.in/nullstone-io/nullstone.v0/generic"
	"log"
)

const (
	ClusterModuleType = "cluster/aws-fargate"
)

var _ deploy.InfraConfig = InfraConfig{}

// InfraConfig provides the mechanism through which AWS actions are performed
type InfraConfig struct {
	BucketName string
	AwsConfig  aws.Config
}

func newInfraConfig(nsConfig api.Config, workspace *types.Workspace) (*InfraConfig, error) {
	dc := &InfraConfig{}
	missingErr := generic.ErrMissingOutputs{OutputNames: []string{}}

	if workspace.LastSuccessfulRun == nil || workspace.LastSuccessfulRun.Apply == nil {
		return nil, fmt.Errorf("cannot find outputs for application")
	}
	workspaceOutputs := workspace.LastSuccessfulRun.Apply.Outputs
	if dc.BucketName = generic.ExtractStringFromOutputs(workspaceOutputs, "bucket_name"); dc.BucketName == "" {
		missingErr.OutputNames = append(missingErr.OutputNames, "bucket_name")
	}

	deployerUser := nsaws.DeployerUser{}
	if !generic.ExtractStructFromOutputs(workspaceOutputs, "deployer", &deployerUser) {
		missingErr.OutputNames = append(missingErr.OutputNames, "deployer")
	}
	dc.AwsConfig = deployerUser.CreateConfig()

	if len(missingErr.OutputNames) > 0 {
		return nil, missingErr
	}
	return dc, nil
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger.Printf("Using s3 bucket %q\n", c.BucketName)
}

func (c InfraConfig) SyncArchive(archive string) error {
	panic("not implemented")
}

func (c InfraConfig) SyncDirectory(dir string) error {
	panic("not implemented")
}
