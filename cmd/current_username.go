package cmd

import (
	"context"
	"errors"
	"fmt"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/response"
)

// OrgApiKeyUsername is the attribution name (e.g. NULLSTONE_TRIGGER_NAME, ECS "started by")
// used when the CLI is authenticated with an organization API key.
// Org API keys are not backed by a user, so the API rejects /current_user for them.
const OrgApiKeyUsername = "org-api-key"

// orgApiKeyUnsupportedErrorType is the API error type emitted when an endpoint
// that acts on behalf of a user is called with an organization API key.
const orgApiKeyUnsupportedErrorType = "problems/org-api-key-unsupported"

// getCurrentUsername resolves a display name for the current credentials to attribute run/exec commands.
// User credentials (including personal API keys) resolve to the user's name.
// Organization API keys have no backing user; they fall back to OrgApiKeyUsername.
func getCurrentUsername(ctx context.Context, cfg api.Config) (string, error) {
	client := api.Client{Config: cfg}
	user, err := client.CurrentUser().Get(ctx)
	if err != nil {
		var unauthorizedErr response.UnauthorizedError
		if errors.As(err, &unauthorizedErr) && unauthorizedErr.Type == orgApiKeyUnsupportedErrorType {
			return OrgApiKeyUsername, nil
		}
		return "", fmt.Errorf("unable to fetch the current user: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("unable to load the current user info")
	}
	return user.Name, nil
}
