package main

import (
	"os"
	"runtime"

	"github.com/satshub/go-bitcoind/address"
	"github.com/satshub/go-bitcoind/usecase/address/config"
	"github.com/satshub/go-bitcoind/usecase/address/config/utils"
	"github.com/satshub/go-bitcoind/usecase/address/log"
	"github.com/urfave/cli"
)

func initLog() {
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)
}

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Ontology CLI"
	app.Action = startService
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{
		Bip39Executive,
		Bip44Executive,
		Bip49Executive,
		Bip84Executive,
	}
	app.Flags = []cli.Flag{
		utils.ConfigFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		initLog()
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

func startService(ctx *cli.Context) {
	address.DefaultExecutive()
}
