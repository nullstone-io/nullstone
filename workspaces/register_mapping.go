package workspaces

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/response"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// registerMappingTimeout bounds the mapping call so that selecting a workspace is never held up by it.
const registerMappingTimeout = 5 * time.Second

// RegisterMapping tells the state backend which workspace (stack/block/env) the selected
// state workspace stores state for. Nullstone scopes platform data to a stack through it.
//
// The state backend creates its workspace the first time terraform reads it, so this runs after `init`.
// The server inserts the mapping when absent and does nothing when present.
//
// This is best effort and never fails the command: a failure is reported once on stderr.
// A state workspace that does not exist yet (404) is not reported; the next select or run registers it.
func RegisterMapping(ctx context.Context, cfg api.Config, workspace Manifest, stderr io.Writer) {
	workspaceUid, err := uuid.Parse(workspace.WorkspaceUid)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, registerMappingTimeout)
	defer cancel()

	client := api.Client{Config: cfg}
	input := types.WorkspaceMappingInput{
		BlockId:      workspace.BlockId,
		EnvId:        workspace.EnvId,
		WorkspaceUid: workspaceUid,
	}
	if _, err := client.WorkspaceMappings().Put(ctx, workspace.StackId, workspaceUid, input); err != nil {
		if response.IsNotFoundError(err) {
			return
		}
		fmt.Fprintf(stderr, "Warning: unable to register the workspace with the Nullstone state backend: %s\n", err)
		fmt.Fprintln(stderr, "The workspace is selected and ready to use.")
	}
}
