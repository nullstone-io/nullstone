package nsaws

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	caws "gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
)

func NewCloudfrontClient(user caws.User, region string) *cloudfront.Client {
	cfg := NewConfig(user, region)
	opts := cloudfront.Options{
		Region:        cfg.Region,
		HTTPClient:    cfg.HTTPClient,
		Credentials:   cfg.Credentials,
		APIOptions:    cfg.APIOptions,
		Logger:        cfg.Logger,
		ClientLogMode: cfg.ClientLogMode,
	}
	if cfg.Retryer != nil {
		opts.Retryer = cfg.Retryer()
	}
	return cloudfront.New(opts)
}
