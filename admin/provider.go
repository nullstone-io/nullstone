package admin

import (
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type NewStatuserFunc func(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (Statuser, error)
type NewRemoterFunc func(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (Remoter, error)
type NewLogStreamerFunc func(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (LogStreamer, error)

type Provider struct {
	NewStatuser    NewStatuserFunc
	NewRemoter     NewRemoterFunc
	NewLogStreamer NewLogStreamerFunc
}
