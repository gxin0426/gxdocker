package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `mydocker`

func main() {

	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	fmt.Println("yigongjici")
	app.Commands = []cli.Command{

		runCommand,
		commitCommand,
		initCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
