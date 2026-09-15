package cmd

import (
	"testing"

	"github.com/nullstone-io/module/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestConnectionScope_Header(t *testing.T) {
	tests := []struct {
		name  string
		scope connectionScope
		want  string
	}{
		{
			name:  "root module",
			scope: connectionScope{},
			want:  "There are connections in this module that do not have a target set.",
		},
		{
			name:  "capability with module and version",
			scope: connectionScope{CapabilityName: "ingress", ModuleSource: "nullstone/aws-ecs-ingress", ModuleVersion: "0.4.2"},
			want:  `Capability "ingress" (nullstone/aws-ecs-ingress@0.4.2) has connections that do not have a target set.`,
		},
		{
			name:  "capability with module but no resolved version",
			scope: connectionScope{CapabilityName: "ingress", ModuleSource: "nullstone/aws-ecs-ingress"},
			want:  `Capability "ingress" (nullstone/aws-ecs-ingress) has connections that do not have a target set.`,
		},
		{
			name:  "capability with no module",
			scope: connectionScope{CapabilityName: "ingress"},
			want:  `Capability "ingress" has connections that do not have a target set.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.scope.Header())
		})
	}
}

func TestConnectionScope_Prompt(t *testing.T) {
	tests := []struct {
		name  string
		scope connectionScope
		conn  types.Connection
		want  string
	}{
		{
			name:  "root required with contract",
			scope: connectionScope{},
			conn:  types.Connection{Connection: config.Connection{Contract: "network/aws/vpc"}},
			want:  `[required] connection "network" (contract=network/aws/vpc):`,
		},
		{
			name:  "capability optional with contract",
			scope: connectionScope{CapabilityName: "ingress", ModuleSource: "nullstone/aws-ecs-ingress"},
			conn:  types.Connection{Connection: config.Connection{Contract: "cluster/aws/ecs:*", Optional: true}},
			want:  `[optional] ingress → connection "network" (contract=cluster/aws/ecs:*):`,
		},
		{
			name:  "falls back to deprecated type when contract is empty",
			scope: connectionScope{},
			conn:  types.Connection{Connection: config.Connection{Type: "cluster/aws-ecs"}},
			want:  `[required] connection "network" (type=cluster/aws-ecs):`,
		},
		{
			name:  "omits schema when both contract and type are empty",
			scope: connectionScope{CapabilityName: "ingress"},
			conn:  types.Connection{},
			want:  `[required] ingress → connection "network":`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.scope.Prompt("network", tt.conn))
		})
	}
}
