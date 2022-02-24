package workspaces

import "github.com/nullstone-io/module/config"

// ScanLocal scans the working directory to build a module manifest
// This is useful if a user is actively making changes to a module that have not been published
func ScanLocal(dir string) (*config.Manifest, error) {
	tfconfig, err := config.ParseDir(".")
	if err != nil {
		return nil, err
	}
	manifest := tfconfig.ToManifest()
	return &manifest, nil
}
