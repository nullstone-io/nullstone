package cmd

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
)

// iacSyncOwnershipHint explains what to do when an IaC sync failed because the repository it
// ran for does not own the blocks or events it defines: the recorded owner may be stale (the
// owning repo was disconnected before a sync could release them), and a stack owner or
// architect can change it from the settings pages linked here. The failure message itself is
// printed unchanged; this is an extra paragraph after it. It returns "" for any other failure.
//
// blocks is the stack's block list, used to turn conflicting block names into settings urls;
// a name with no block (or a nil list when the lookup failed) is printed without a url.
func iacSyncOwnershipHint(cfg api.Config, iw types.IntentWorkflow, blocks []types.Block) string {
	conflict := types.MatchIacOwnershipConflict(iw.StatusMessage)
	if conflict == nil {
		return ""
	}
	byName := map[string]types.Block{}
	for _, block := range blocks {
		byName[block.Name] = block
	}

	lines := []string{
		"If this ownership is out of date (for example the repository was disconnected before a sync could release these),",
		"a stack owner or architect can change the owning repository in the Nullstone UI:",
	}
	for _, name := range conflict.BlockNames() {
		if block, ok := byName[name]; ok {
			lines = append(lines, fmt.Sprintf("  %s: %s", name, app_urls.GetBlockIacOwnershipSettings(cfg, block)))
		} else {
			lines = append(lines, fmt.Sprintf("  %s: block settings > IaC ownership", name))
		}
	}
	if names := conflict.EventNames(); len(names) > 0 {
		lines = append(lines, fmt.Sprintf("  events (%s): %s", strings.Join(names, ", "),
			app_urls.GetEnvIacOwnershipSettings(cfg, iw.OrgName, iw.StackId, iw.EnvId)))
	}
	return strings.Join(lines, "\n")
}

// iacSyncFailureMessage is the exit message for a failed IaC sync: the workflow's status message
// as-is, followed by the ownership hint when one applies. The block lookup is best effort; a
// failure there only costs the per-block urls.
func iacSyncFailureMessage(ctx context.Context, cfg api.Config, iw types.IntentWorkflow) string {
	msg := fmt.Sprintf("IaC sync failed: %s", iw.StatusMessage)
	if types.MatchIacOwnershipConflict(iw.StatusMessage) == nil {
		return msg
	}
	client := api.Client{Config: cfg}
	blocks, _ := client.Blocks().List(ctx, iw.StackId, false)
	return msg + "\n\n" + iacSyncOwnershipHint(cfg, iw, blocks)
}
