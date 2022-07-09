package aws_ecs_ec2

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecr"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecs"
	"log"
)

var ModuleContractName = types.ModuleContractName{
	Category:    string(types.CategoryApp),
	Subcategory: string(types.SubcategoryAppContainer),
	Provider:    "aws",
	Platform:    "ecs",
	Subplatform: "ec2",
}

func NewProvider(logger *log.Logger, nsConfig api.Config, appDetails app.Details) app.Provider {
	return Provider{
		Logger:     logger,
		NsConfig:   nsConfig,
		AppDetails: appDetails,
	}
}

type Provider struct {
	Logger     *log.Logger
	NsConfig   api.Config
	AppDetails app.Details
}

func (p Provider) NewPusher() (app.Pusher, error) {
	return ecr.NewPusher(p.Logger, p.NsConfig, p.AppDetails)
}

func (p Provider) NewDeployer() (app.Deployer, error) {
	return ecs.NewDeployer(p.Logger, p.NsConfig, p.AppDetails)
}

func (p Provider) NewDeployStatusGetter() (app.DeployStatusGetter, error) {
	//TODO implement me
	panic("implement me")
}
