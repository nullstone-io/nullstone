package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"sort"
)

var Version = "development"

func main() {
	app := &cli.App{
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
