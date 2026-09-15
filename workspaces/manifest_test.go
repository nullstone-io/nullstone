package workspaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestManifestConnectionsFrom(t *testing.T) {
	envId := int64(7)
	conns := types.Connections{
		"network": {
			EffectiveTarget: &types.ConnectionTarget{StackId: 1, BlockId: 10, BlockName: "network0", EnvId: &envId},
		},
		"cluster": {
			EffectiveTarget: &types.ConnectionTarget{StackId: 1, BlockId: 11, BlockName: "cluster0"},
		},
		"unresolved": {
			EffectiveTarget: nil,
		},
		"zero-block": {
			EffectiveTarget: &types.ConnectionTarget{StackId: 1, BlockId: 0, BlockName: ""},
		},
	}

	got := ManifestConnectionsFrom(conns)

	assert.Equal(t, ManifestConnections{
		"network": {StackId: 1, BlockId: 10, BlockName: "network0", EnvId: &envId},
		"cluster": {StackId: 1, BlockId: 11, BlockName: "cluster0"},
	}, got)
}
