package main

import (
	"github.com/codegangsta/cli"
	"github.com/ulrf/ulrf/torefactor"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "orgs"
	app.Usage = "orgs website"
	app.Action = torefactor.RunMacaron

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "mode, m",
			Value: "dev",
			Usage: "mode dev|prod",
		},
	}

	app.Run(os.Args)
}
