package config

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	ApiKeyFilename string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to find home directory: %s\n", err)
	}
	ApiKeyFilename = path.Join(home, ".nullstone", "api-key")
}

func SaveApiKey(apiKey string) error {
	if err := os.MkdirAll(filepath.Dir(ApiKeyFilename), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(ApiKeyFilename, []byte(apiKey), 0644)
}

func ReadApiKey() (string, error) {
	if _, err := os.Stat(ApiKeyFilename); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	raw, err := ioutil.ReadFile(ApiKeyFilename)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
