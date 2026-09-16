package cmd

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestIacSyncOwnershipHint(t *testing.T) {
	cfg := api.Config{BaseAddress: "https://api.nullstone.io", OrgName: "acme"}
	iw := types.IntentWorkflow{OrgName: "acme", StackId: 12, EnvId: 56}
	blocks := []types.Block{
		{IdModel: types.IdModel{Id: 34}, OrgName: "acme", StackId: 12, Type: "Application", Name: "api"},
		{IdModel: types.IdModel{Id: 35}, OrgName: "acme", StackId: 12, Type: "Datastore", Name: "db"},
		{IdModel: types.IdModel{Id: 36}, OrgName: "acme", StackId: 12, Type: "Subdomain", Name: "www"},
	}
	events := []types.EnvEvent{
		{Uid: uuid.MustParse("0f9c2d4e-3b1a-4c6d-8e7f-a1b2c3d4e5f6"), OrgName: "acme", StackId: 12, EnvId: 56, Name: "deploy-notify"},
	}
	// verbatim nullfire messages (internal/blocks/errors.go, iac/update_events.go)
	blockConflict := strings.Join([]string{
		"IaC sync from acme/thief cannot update 3 blocks owned by other repositories:",
		"\tacme/infra owns api, www",
		"\tacme/other owns gone",
		"A block can only be defined in one repository for IaC sync, remove it from one of them.",
	}, "\n")
	eventConflict := "2 errors occurred:\n" +
		"\t* event \"deploy-notify\" is configured in another repository (https://github.com/acme/infra)\n" +
		"\t* event \"approvals\" is configured in another repository (https://github.com/acme/other)\n\n"
	intro := []string{
		"If this ownership is out of date (for example the repository was disconnected before a sync could release these),",
		"a stack owner or architect can change the owning repository in the Nullstone UI:",
	}
	withIntro := func(lines ...string) []string {
		return append(append([]string{}, intro...), lines...)
	}

	tests := []struct {
		name   string
		status string
		blocks []types.Block
		events []types.EnvEvent
		want   []string
	}{
		{
			name:   "not an ownership conflict",
			status: "An error occurred when validating Nullstone IaC files.\nblock \"api\": bad",
			blocks: blocks,
			events: events,
			want:   nil,
		},
		{
			name:   "blocks link to their settings by type; an unknown name is listed without a url",
			status: blockConflict,
			blocks: blocks,
			want: withIntro(
				"  block api: https://app.nullstone.io/orgs/acme/stacks/12/apps/34/settings#iac-ownership",
				"  block gone: block settings > IaC ownership",
				"  block www: https://app.nullstone.io/orgs/acme/stacks/12/blocks/36/settings#iac-ownership",
			),
		},
		{
			name:   "block lookup failed: names without urls",
			status: blockConflict,
			blocks: nil,
			want: withIntro(
				"  block api: block settings > IaC ownership",
				"  block gone: block settings > IaC ownership",
				"  block www: block settings > IaC ownership",
			),
		},
		{
			name:   "events link to their edit page; an unknown name points at the events list",
			status: eventConflict,
			events: events,
			want: withIntro(
				"  event approvals: https://app.nullstone.io/orgs/acme/stacks/12/envs/56/events > edit > IaC ownership",
				"  event deploy-notify: https://app.nullstone.io/orgs/acme/stacks/12/envs/56/events/0f9c2d4e-3b1a-4c6d-8e7f-a1b2c3d4e5f6#iac-ownership",
			),
		},
		{
			name:   "event lookup failed: every event points at the events list",
			status: eventConflict,
			events: nil,
			want: withIntro(
				"  event approvals: https://app.nullstone.io/orgs/acme/stacks/12/envs/56/events > edit > IaC ownership",
				"  event deploy-notify: https://app.nullstone.io/orgs/acme/stacks/12/envs/56/events > edit > IaC ownership",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cur := iw
			cur.StatusMessage = tt.status
			got := iacSyncOwnershipHint(cfg, cur, tt.blocks, tt.events)
			if tt.want == nil {
				assert.Equal(t, "", got)
				return
			}
			assert.Equal(t, strings.Join(tt.want, "\n"), got)
		})
	}
}
