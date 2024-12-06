package beanstalk

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/creds"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	Region        string     `ns:"region"`
	BeanstalkArn  string     `ns:"beanstalk_arn"`
	EnvironmentId string     `ns:"environment_id"`
	Adminer       nsaws.User `ns:"adminer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *types.Workspace) {
	credsFactory := creds.NewProviderFactory(source, ws.StackId, ws.Uid)
	o.Adminer.RemoteProvider = credsFactory("adminer")
}

func (o *Outputs) AdminerConfig() aws.Config {
	return nsaws.NewConfig(o.Adminer, o.Region)
}
