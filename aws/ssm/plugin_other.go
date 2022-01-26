//go:build !darwin && !windows
// +build !darwin,!windows

package ssm

func getSessionManagerPluginPath() (string, error) {
	return "session-manager-plugin", nil
}
