package workspaces

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func TestRegisterMapping(t *testing.T) {
	workspaceUid := "6f1a2b3c-4d5e-4f60-8a9b-0c1d2e3f4a5b"

	tests := []struct {
		name         string
		workspaceUid string
		status       int
		body         string
		wantCalls    int32
		wantWarning  bool
	}{
		{
			name:         "registers the mapping of the selected workspace",
			workspaceUid: workspaceUid,
			status:       http.StatusOK,
			body:         `{"stackId": 42, "blockId": 7, "envId": 9}`,
			wantCalls:    1,
		},
		{
			name:         "is silent when the state workspace does not exist yet",
			workspaceUid: workspaceUid,
			status:       http.StatusNotFound,
			body:         `{"message": "Workspace not found"}`,
			wantCalls:    1,
		},
		{
			name:         "warns when the mapping is rejected",
			workspaceUid: workspaceUid,
			status:       http.StatusForbidden,
			body:         `{"message": "mismatch"}`,
			wantCalls:    1,
			wantWarning:  true,
		},
		{
			name:         "warns when verification fails",
			workspaceUid: workspaceUid,
			status:       http.StatusBadGateway,
			body:         `{"message": "Unable to verify the workspace with Nullstone."}`,
			wantCalls:    1,
			wantWarning:  true,
		},
		{
			name:         "skips a workspace without a uid",
			workspaceUid: "",
			status:       http.StatusOK,
			wantCalls:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			var gotMethod, gotPath string
			var gotBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				gotMethod, gotPath = r.Method, r.URL.Path
				gotBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			t.Cleanup(server.Close)
			cfg := api.Config{BaseAddress: server.URL, OrgName: "nullstone"}
			manifest := Manifest{OrgName: "nullstone", StackId: 42, BlockId: 7, EnvId: 9, WorkspaceUid: test.workspaceUid}

			stderr := &bytes.Buffer{}
			RegisterMapping(context.Background(), cfg, manifest, stderr)

			assert.Equal(t, test.wantCalls, calls.Load())
			if test.wantCalls > 0 {
				assert.Equal(t, http.MethodPut, gotMethod)
				assert.Equal(t, "/orgs/nullstone/stacks/42/workspaces/"+workspaceUid+"/mapping", gotPath)
				assert.JSONEq(t, `{"blockId": 7, "envId": 9, "workspaceUid": "`+workspaceUid+`"}`, string(gotBody))
			}
			if test.wantWarning {
				assert.Equal(t, 1, bytes.Count(stderr.Bytes(), []byte("Warning:")), "expected exactly one warning, got %q", stderr.String())
			} else {
				assert.Empty(t, stderr.String())
			}
		})
	}
}
