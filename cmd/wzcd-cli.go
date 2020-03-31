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

	// Setup MQ
	controller.GetTransport().AddNatsServerURL(
		conf.Find("transport").String("host", ""),
		conf.Find("transport").DefaultInt("port", "", 4222),
	)

	// Setup DB
	confDb := conf.Find("db")
	controller.GetDb().SetHost(confDb.String("host", "")).
		SetPort(confDb.DefaultInt("port", "", 26257)).
		SetUser(confDb.String("user", "")).
		SetDbName(confDb.String("database", "")).
		SetSSLConf(confDb.String("ssl_root", ""),
			confDb.String("ssl_key", ""),
			confDb.String("ssl_cert", ""))

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
