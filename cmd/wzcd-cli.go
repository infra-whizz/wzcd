package main

import (
	"os"

	wzlib_utils "github.com/infra-whizz/wzlib/utils"

	"github.com/infra-whizz/wzcd"
	"github.com/isbm/go-nanoconf"
	"github.com/urfave/cli/v2"
)

func setupControllerInstance(ctx *cli.Context, level logrus.Level) *wzcd.WzcDaemon {
	conf := nanoconf.NewConfig(ctx.String("config"))
	controller := wzcd.NewWzcDaemon()
	controller.GetLogger().SetLevel(level)

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

	controller.SetupMachineIdUtil("") // read-only mode, as it points to a default D-Bus file.

	return controller
}

func b2i(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

// PKI manager function
func appManagePKI(ctx *cli.Context) error {
	controller := setupControllerInstance(ctx)
	controller.GetDb().Open()
	defer controller.GetDb().Close()

	if b2i(ctx.Bool("list"))+b2i(ctx.Bool("add"))+b2i(ctx.Bool("remove")) != 1 {
		controller.GetLogger().Errorln("You can only list or add or remove at a time.")
		os.Exit(wzlib_utils.EX_USAGE)
	}

	if ctx.Bool("list") {
		controller.GetPKIManager().ListRemotePEMKeys()
	} else if ctx.Bool("add") {
		if err := controller.GetPKIManager().RegisterPEMKey(ctx.String("path"), ctx.String("machineid"), ctx.String("fqdn")); err != nil {
			controller.GetLogger().Errorln(err.Error())
		}
	} else if ctx.Bool("remove") {
		fingerprint := ctx.String("fingerprint")
		if err := controller.GetPKIManager().RemovePEMKey(fingerprint); err != nil {
			controller.GetLogger().Errorln(err.Error())
		}
	}

	return nil
}

// Main runner function
func appMainRun(ctx *cli.Context) error {
	setupControllerInstance(ctx, logrus.DebugLevel).Run().AppLoop()
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
		Action:  appMainRun,
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
			Name:   "pki",
			Usage:  "Manage PKI of Whizz remotes",
			Action: appManagePKI,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "list",
					Usage:   "List all public keys for all available remotes",
					Aliases: []string{"l"},
				},
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
