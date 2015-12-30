package main

import (
	"github.com/codegangsta/cli"
	"github.com/ulrf/ulrf/torefactor"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
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

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Println(mem.Alloc)
	log.Println(mem.TotalAlloc)
	log.Println(mem.HeapAlloc)
	log.Println(mem.HeapSys)

	app.Run(os.Args)
}
