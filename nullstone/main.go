package main

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cliApp := app.Build()
	cliApp.Version = version
	cliApp.Metadata = map[string]interface{}{
		"commit":  commit,
		"date":    date,
		"builtBy": builtBy,
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
