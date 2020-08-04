package main

import (
	"fmt"
	"gxdocker/cgroups/subsystems"
	"gxdocker/container"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `create a container with namespace and cgroup limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},

		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},

		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory list",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},

		cli.StringFlag{
			Name:  "name",
			Usage: "con name",
		},
	},
	Action: func(context *cli.Context) error {

		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}

		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		fmt.Println(cmdArray)
		tty := context.Bool("ti")

		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("ti and d param can not both provided")
		}

		volume := context.String("v")

		containerName := context.String("name")

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		fmt.Println(tty, cmdArray, resConf, volume, containerName)
		Run(tty, cmdArray, resConf, volume, containerName)
		return nil

	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}

		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all con",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}
