package main

import (
	"os"
	"sort"

	"github.com/mintxtinm/packet-monitor/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {

	app := cli.NewApp()
	app.Name = "Packet Monitor"
	app.Description = "Packet Monitor"
	app.Usage = ""
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "server",
			Value: true,
			Usage: "Exported file server",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "xdp",
			Usage: "XDP dependencies graph toolkit",
			Subcommands: []*cli.Command{
				cmd.LoadXdpTracerCommand,
				cmd.UnloadXdpTracerCommand,
			},
		},
	}
	app.Flags = []cli.Flag{}

	app.Before = func(c *cli.Context) (err error) {
		err = os.MkdirAll("./export", os.ModePerm)
		return
	}

	// sort all flags
	for _, cmd := range app.Commands {
		sort.Sort(cli.FlagsByName(cmd.Flags))
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

}
