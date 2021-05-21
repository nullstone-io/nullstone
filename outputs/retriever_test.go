package outputs

import (
	"github.com/nullstone-io/module/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"testing"
)

type MockFlatOutputs struct {
	Output1 string            `ns:"output1"`
	Output2 int               `ns:"output2"`
	Output3 map[string]string `ns:"output3"`
}

type MockDeepOutputs struct {
	Output1    string          `ns:"output1"`
	Connection MockFlatOutputs `ns:",connectionType:aws-flat"`
}

func TestRetriever_Retrieve(t *testing.T) {
	flatWorkspace := &types.Workspace{
		OrgName: "default",
		StackId: 1,
		BlockId: 5,
		EnvId:   15,
		LastSuccessfulRun: &types.Run{
			Apply: &types.RunApply{
				Outputs: types.Outputs{
					"output1": types.OutputItem{
						Type:      "string",
						Value:     "value1",
						Sensitive: false,
					},
					"output2": types.OutputItem{
						Type:      "number",
						Value:     2,
						Sensitive: false,
					},
					"output3": types.OutputItem{
						Type: "map(string)",
						Value: map[string]string{
							"key1": "value1",
							"key2": "value2",
							"key3": "value3",
						},
						Sensitive: false,
					},
				},
			},
		},
	}
	deepWorkspace := &types.Workspace{
		OrgName: "default",
		StackId: 1,
		BlockId: 6,
		EnvId:   15,
		LastSuccessfulRun: &types.Run{
			Config: &types.RunConfig{
				Connections: map[string]types.Connection{
					"deep": {
						Connection: config.Connection{
							Type:     "aws-flat",
							Optional: false,
						},
						Target: "deep0",
						Reference: &types.BlockConnection{
							StackId: 1,
							BlockId: 5,
							EnvId:   nil,
						},
						Unused: false,
					},
				},
			},
			Apply: &types.RunApply{
				Outputs: types.Outputs{
					"output1": types.OutputItem{
						Type:      "string",
						Value:     "test",
						Sensitive: false,
					},
				},
			},
		},
	}

	t.Run("should retrieve outputs for single workspace", func(t *testing.T) {
		want := MockFlatOutputs{
			Output1: "value1",
			Output2: 2,
			Output3: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		}

		retriever := Retriever{NsConfig: api.Config{}}
		var got MockFlatOutputs
		if assert.NoError(t, retriever.Retrieve(flatWorkspace, &got)) {
			assert.Equal(t, want, got)
		}
	})

	t.Run("should retrieve outputs for own workspace and connected workspace", func(t *testing.T) {
		server, nsConfig := mockNs([]types.Workspace{
			*deepWorkspace,
			*flatWorkspace,
		})
		t.Cleanup(server.Close)

		want := MockDeepOutputs{
			Output1: "test",
			Connection: MockFlatOutputs{
				Output1: "value1",
				Output2: 2,
				Output3: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		}

		retriever := Retriever{NsConfig: nsConfig}
		var got MockDeepOutputs
		if assert.NoError(t, retriever.Retrieve(deepWorkspace, &got)) {
			assert.Equal(t, want, got)
		}
	})
}
