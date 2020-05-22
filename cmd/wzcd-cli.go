package main

import (
	"fmt"
	"os"
	"path"

	"github.com/infra-whizz/wzcd"
	wzlib_utils "github.com/infra-whizz/wzlib/utils"
	"github.com/isbm/go-nanoconf"
	"github.com/sirupsen/logrus"
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

	controller.SetPKIDir(conf.Root().String("pki", ""))
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

// Manage Cluster PKI
func appManageClusterPKI(ctx *cli.Context) error {
	controller := setupControllerInstance(ctx, logrus.DebugLevel)
	controller.GetDb().Open()
	defer controller.GetDb().Close()

	if ctx.Bool("show") {
		clusterFingerprint := controller.GetPKIManager().GetClusterPublicPEMKeyFingerprint()
		if clusterFingerprint != "" {
			fmt.Println("Cluster fingerprint:", clusterFingerprint)
		} else {
			fmt.Println("No cluster PKI has been yet defined. Please add one!")
			os.Exit(wzlib_utils.EX_UNAVAILABLE)
		}
	} else if ctx.Bool("rotate-rsa-keys") {
		if ctx.String("public") != "" || ctx.String("private") != "" {
			if ctx.String("public") == "" || ctx.String("private") == "" {
				fmt.Println("Unable to set RSA keypair: should be defined both public and private key.")
				os.Exit(wzlib_utils.EX_USAGE)
			} else {
				fmt.Println("Set RSA keypair")
				if err := controller.GetPKIManager().RegisterClusterPEMKeyPair(ctx.String("public"), ctx.String("private")); err != nil {
					fmt.Println("Error registering PEM pair:", err.Error())
				}
			}
		} else {
			fmt.Println("Rotate with automatic pre-generation")
			if err := controller.GetCryptoBundle().GetRSA().GenerateKeyPair(controller.GetPKIDir()); err != nil {
				fmt.Println("Error RSA keypair rotation:", err.Error())
			}
			pubkeyPath := path.Join(controller.GetPKIDir(), "public.pem")
			privkeyPath := path.Join(controller.GetPKIDir(), "private.pem")
			if err := controller.GetPKIManager().RegisterClusterPEMKeyPair(pubkeyPath, privkeyPath); err != nil {
				fmt.Println("Error rotating PEM pair:", err.Error())
			}
		}
	} else {
		if err := cli.ShowSubcommandHelp(ctx); err != nil {
			fmt.Println("Error:", err.Error())
		}
		os.Exit(wzlib_utils.EX_USAGE)
	}

	return nil
}

// Manage remote PKI
func appManageRemotePKI(ctx *cli.Context) error {
	controller := setupControllerInstance(ctx, logrus.DebugLevel)
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
			Name:   "cluster",
			Usage:  "Manage PKI for the entire Whizz cluster",
			Action: appManageClusterPKI,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "show",
					Usage:   "Show cluster RSA keypair (default)",
					Aliases: []string{"l"},
				},
				&cli.BoolFlag{
					Name: "rotate-rsa-keys",
					Usage: "Rotate cluster RSA keypair. Keys wil be re-generated into a default directory. " +
						"\n\tSpecify 'public' and 'private' for alternative paths." +
						"\n\tWARNING: This operation will require update RSA public \n\tkeys on all remotes!",
				},
				&cli.StringFlag{
					Name:    "public",
					Usage:   "Path to the public key in PEM format for the cluster",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "private",
					Usage:   "Path to the private key in PEM format of the cluster",
					Aliases: []string{"i"},
				},
			},
		},
		{
			Name:   "remote",
			Usage:  "Manage PKI of Whizz remotes",
			Action: appManageRemotePKI,
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
