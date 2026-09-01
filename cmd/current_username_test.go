package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func TestGetCurrentUsername(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:   "user credentials resolve to the user's name",
			status: http.StatusOK,
			body:   `{"id":1,"name":"brad.sickles","picture":""}`,
			want:   "brad.sickles",
		},
		{
			name:   "org api key falls back to the org-api-key name",
			status: http.StatusForbidden,
			body:   `{"type":"problems/org-api-key-unsupported","title":"Unsupported for Organization API Key","code":403,"message":"This endpoint cannot be used with an organization API key."}`,
			want:   OrgApiKeyUsername,
		},
		{
			name:        "other 403 errors still fail",
			status:      http.StatusForbidden,
			body:        `{"type":"problems/authorization-error","title":"Access Denied","code":403,"message":"You do not have the proper authorization to access this resource."}`,
			wantErr:     true,
			errContains: "unable to fetch the current user",
		},
		{
			name:        "401 errors still fail",
			status:      http.StatusUnauthorized,
			body:        `{"type":"problems/authentication-error","title":"Not Authenticated","code":401,"message":"You must login to access this resource."}`,
			wantErr:     true,
			errContains: "unable to fetch the current user",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/current_user", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			}))
			defer server.Close()

			cfg := api.Config{BaseAddress: server.URL}
			got, err := getCurrentUsername(context.Background(), cfg)
			if test.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
