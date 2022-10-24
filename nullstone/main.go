package main

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"os"
)

func main() {
	cliApp := app.Build()

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
