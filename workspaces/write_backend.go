package workspaces

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
	"path/filepath"
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

func WriteBackendTf(cfg api.Config, workspaceUid uuid.UUID, filename string) error {
	backend := fmt.Sprintf(backendTmpl, cfg.BaseAddress, cfg.OrgName, workspaceUid)
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(backend), 0644)
}
