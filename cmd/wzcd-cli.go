package main

import (
	"os"

	"github.com/infra-whizz/wzcd"
	"github.com/isbm/go-nanoconf"
	"github.com/urfave/cli/v2"
)

func manageWhizz(ctx *cli.Context) error {
	return nil
}

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
	app.Commands = []*cli.Command{
		{
			Name:   "whizz",
			Usage:  "Manage Whizz remotes",
			Action: manageWhizz,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "add",
					Usage:   "Add public key in PEM format",
					Aliases: []string{"a"},
				},
				&cli.BoolFlag{
					Name:    "remove",
					Usage:   "Remove public key",
					Aliases: []string{"r", "d"},
				},
				&cli.StringFlag{
					Name:    "path",
					Usage:   "Path to the file",
					Aliases: []string{"p"},
				},
				&cli.StringFlag{
					Name:    "fingerprint",
					Usage:   "Fingerprint of the PEM key",
					Aliases: []string{"f"},
				},
				&cli.StringFlag{
					Name:    "machineid",
					Usage:   "Machine ID",
					Aliases: []string{"i"},
				},
				&cli.StringFlag{
					Name:    "fqdn",
					Usage:   "FQDN of the Whizz host",
					Aliases: []string{"n"},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
