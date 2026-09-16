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
// architect can change it from the "IaC ownership" section on each block's settings page or
// each event's edit page. The failure message itself is printed unchanged; this is an extra
// paragraph after it. It returns "" for any other failure.
//
// blocks and events are the stack's blocks and the env's events, used to turn conflicting
// names into urls; a name with no match (or a nil list when the lookup failed) gets a generic
// pointer instead.
func iacSyncOwnershipHint(cfg api.Config, iw types.IntentWorkflow, blocks []types.Block, events []types.EnvEvent) string {
	conflict := types.MatchIacOwnershipConflict(iw.StatusMessage)
	if conflict == nil {
		return ""
	}
	blockByName := map[string]types.Block{}
	for _, block := range blocks {
		blockByName[block.Name] = block
	}
	eventByName := map[string]types.EnvEvent{}
	for _, event := range events {
		eventByName[event.Name] = event
	}

	lines := []string{
		"If this ownership is out of date (for example the repository was disconnected before a sync could release these),",
		"a stack owner or architect can change the owning repository in the Nullstone UI:",
	}
	for _, name := range conflict.BlockNames() {
		if block, ok := blockByName[name]; ok {
			lines = append(lines, fmt.Sprintf("  block %s: %s", name, app_urls.GetBlockIacOwnershipSettings(cfg, block)))
		} else {
			lines = append(lines, fmt.Sprintf("  block %s: block settings > IaC ownership", name))
		}
	}
	for _, name := range conflict.EventNames() {
		if event, ok := eventByName[name]; ok {
			lines = append(lines, fmt.Sprintf("  event %s: %s", name, app_urls.GetEnvEventIacOwnership(cfg, event)))
		} else {
			lines = append(lines, fmt.Sprintf("  event %s: %s > edit > IaC ownership", name, app_urls.GetEnvEvents(cfg, iw.OrgName, iw.StackId, iw.EnvId)))
		}
	}
	return strings.Join(lines, "\n")
}

// iacSyncFailureMessage is the exit message for a failed IaC sync: the workflow's status message
// as-is, followed by the ownership hint when one applies. The block and event lookups are best
// effort; a failure there only costs the per-item urls.
func iacSyncFailureMessage(ctx context.Context, cfg api.Config, iw types.IntentWorkflow) string {
	msg := fmt.Sprintf("IaC sync failed: %s", iw.StatusMessage)
	conflict := types.MatchIacOwnershipConflict(iw.StatusMessage)
	if conflict == nil {
		return msg
	}
	client := api.Client{Config: cfg}
	var blocks []types.Block
	if len(conflict.Blocks) > 0 {
		blocks, _ = client.Blocks().List(ctx, iw.StackId, false)
	}
	var events []types.EnvEvent
	if len(conflict.Events) > 0 {
		events, _ = client.EnvEvents().List(ctx, iw.StackId, iw.EnvId)
	}
	return msg + "\n\n" + iacSyncOwnershipHint(cfg, iw, blocks, events)
}
