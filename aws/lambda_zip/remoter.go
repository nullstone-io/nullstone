package lambda_zip

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	lambda_zip "github.com/nullstone-io/deployment-sdk/aws/lambda-zip"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/aws/lambda"
	"strings"
)

func NewRemoter(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[lambda_zip.Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	return Remoter{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type Remoter struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     lambda_zip.Outputs
}

func (r Remoter) Exec(ctx context.Context, options admin.RemoteOptions, cmd []string) error {
	return lambda.Invoke(ctx, r.Infra, []byte(strings.Join(cmd, "")), options.Async)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	return fmt.Errorf("ssh for lambda is not supported")
}
