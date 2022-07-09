package outputs

import (
	"encoding/json"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
	"net/http"
	"net/http/httptest"
)

func mockNs(workspaces []types.Workspace) (*httptest.Server, api.Config) {
	mux := http.NewServeMux()
	for _, workspace := range workspaces {
		cur := workspace
		endpoint := fmt.Sprintf("/orgs/%s/stacks/%d/blocks/%d/envs/%d",
			cur.OrgName, cur.StackId, cur.BlockId, cur.EnvId)
		mux.Handle(endpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := json.Marshal(cur)
			w.Write(raw)
		}))
		outputsEndpoint := fmt.Sprintf("/orgs/%s/stacks/%d/blocks/%d/envs/%d/outputs/latest",
			cur.OrgName, cur.StackId, cur.BlockId, cur.EnvId)
		mux.Handle(outputsEndpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := json.Marshal(cur.LastFinishedRun.Apply.Outputs)
			w.Write(raw)
		}))
	}
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("unhandled endpoint in mock nullstone API", r.URL.Path)
		http.NotFound(w, r)
	}))

	server := httptest.NewServer(mux)
	return server, api.Config{
		BaseAddress:    server.URL,
		ApiKey:         "invalid-api-key",
		IsTraceEnabled: false,
		OrgName:        "default",
	}
}
