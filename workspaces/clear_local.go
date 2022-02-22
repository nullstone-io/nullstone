package workspaces

import (
	"os"
)

var (
	tfDir = ".terraform/"
)

func HasLocalConfigured() bool {
	_, err := os.Lstat(tfDir)
	return err == nil
}

func ClearLocalConfiguration() error {
	return os.RemoveAll(tfDir)
}
