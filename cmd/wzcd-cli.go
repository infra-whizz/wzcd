package main

import (
	"os"

	"github.com/infra-whizz/wzcd"
	"github.com/isbm/go-nanoconf"
	"github.com/urfave/cli/v2"
)

func run(ctx *cli.Context) error {
	conf := nanoconf.NewConfig(ctx.String("config"))
	controller := wzcd.NewWzcDaemon()
	controller.GetTransport().AddNatsServerURL(
		conf.Find("transport").String("host", ""),
		conf.Find("transport").DefaultInt("port", "", 4222),
	)
	controller.Run().AppLoop()

	cli.ShowAppHelpAndExit(ctx, 1)
	return nil
}

func main() {
	appname := "wzdc"
	confpath := nanoconf.NewNanoconfFinder(appname).DefaultSetup(nil)

	app := &cli.App{
		Version: "0.1 Alpha",
		Name:    appname,
		Usage:   "Whizz Control Daemon",
		Action:  run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Path to configuration file",
				Required: false,
				Value:    confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault(),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
