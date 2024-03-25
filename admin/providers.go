package admin

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/contract"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Providers map[types.ModuleContractName]Provider

func (s Providers) FindStatuser(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (Statuser, error) {
	factory := s.FindFactory(*appDetails.Module)
	if factory == nil || factory.NewStatuser == nil {
		return nil, nil
	}
	return factory.NewStatuser(ctx, osWriters, source, appDetails)
}

func (s Providers) FindRemoter(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (Remoter, error) {
	factory := s.FindFactory(*appDetails.Module)
	if factory == nil || factory.NewRemoter == nil {
		return nil, nil
	}
	return factory.NewRemoter(ctx, osWriters, source, appDetails)
}

func (s Providers) FindFactory(curModule types.Module) *Provider {
	return contract.FindInRegistrarByModule(s, &curModule)
}
