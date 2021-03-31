package config

import (
	"log"
	"os"
	"path"
)

var (
	NullstoneDir string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to find home directory: %s\n", err)
	}
	NullstoneDir = path.Join(home, ".nullstone")
}
