package admin

import (
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
)

type NewStatuserFunc func(osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (Statuser, error)
type NewRemoterFunc func(osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (Remoter, error)

type Provider struct {
	NewStatuser NewStatuserFunc
	NewRemoter  NewRemoterFunc
}
