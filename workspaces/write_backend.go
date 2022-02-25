package workspaces

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
	"path/filepath"
	"strings"
)

var (
	backendTmpl = `terraform {
  backend "remote" {
    hostname     = %q
    organization = %q
    
    workspaces {
      name = %q
    }
  }
}`
)

func WriteBackendTf(cfg api.Config, workspaceUid string, filename string) error {
	// backend stanza expects a hostname without the scheme -> TF will add `https://` automatically
	hostname := strings.Replace(strings.Replace(cfg.BaseAddress, "https://", "", 1), "http://", "", 1)
	backend := fmt.Sprintf(backendTmpl, hostname, cfg.OrgName, workspaceUid)
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(backend), 0644)
}
