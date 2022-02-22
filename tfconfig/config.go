package tfconfig

import "path/filepath"

func GetCredentialsFilename() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.tfrc"), nil
}
