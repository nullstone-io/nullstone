package tfconfig

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"io/ioutil"
	"os"
	"strings"
)

var (
	credentialsFindTmpl = "credentials %q {"
	credentialsTmpl     = `credentials %q {
  token = %q
}
`
)

func IsCredsConfigured(cfg api.Config) bool {
	credsFilename, err := GetCredentialsFilename()
	if err != nil {
		return false
	}

	hostname := getNullstoneHostname(cfg)
	raw, err := ioutil.ReadFile(credsFilename)
	if err != nil {
		return false
	}
	find := fmt.Sprintf(credentialsFindTmpl, hostname)
	return strings.Contains(string(raw), find)
}

// ConfigCreds configures Terraform with configuration to authenticate Terraform with Nullstone server
// This configuration enables Terraform to:
//   - Configure `backend "remote"` to reach Nullstone state backend
//   - Download private modules from the Nullstone registry
func ConfigCreds(ctx context.Context, cfg api.Config) error {
	credsFilename, err := GetCredentialsFilename()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(credsFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	accessToken, err := cfg.AccessTokenSource.GetAccessToken(ctx, cfg.OrgName)
	if err != nil {
		return err
	}
	_, err = file.WriteString(fmt.Sprintf(credentialsTmpl, getNullstoneHostname(cfg), accessToken))
	return err
}

func getNullstoneHostname(cfg api.Config) string {
	return strings.Replace(strings.Replace(cfg.BaseAddress, "https://", "", 1), "http://", "", 1)
}
