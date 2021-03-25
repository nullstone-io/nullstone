package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"sort"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	app := &cli.App{
		Version: version,
		Metadata: map[string]interface{}{
			"commit":  commit,
			"date":    date,
			"builtBy": builtBy,
		},
		Commands: []cli.Command{},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
