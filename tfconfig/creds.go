package tfconfig

import (
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
//   - Utilize `backend "remote"`
//   - Download private modules in the Nullstone registry
func ConfigCreds(cfg api.Config) error {
	credsFilename, err := GetCredentialsFilename()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(credsFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf(credentialsTmpl, getNullstoneHostname(cfg), cfg.ApiKey))
	return err
}

func getNullstoneHostname(cfg api.Config) string {
	return strings.Replace(strings.Replace(cfg.BaseAddress, "https://", "", 1), "http://", "", 1)
}
