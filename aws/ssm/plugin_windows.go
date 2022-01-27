package ssm

import (
	"os"
	"path/filepath"
)

var (
	osSessionManagerPluginPath string
)

func init() {
	homeDir, _ := os.UserHomeDir()
	absHomeDir, _ := filepath.Abs(homeDir)
	vol := filepath.VolumeName(absHomeDir)
	if vol == "" {
		vol = "C:"
	}
	osSessionManagerPluginPath = filepath.Join(vol, "Program Files", "Amazon", "SessionManagerPlugin", "bin", "session-manager-plugin")
}
