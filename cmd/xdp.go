package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mintxtinm/packet-monitor/pkg/fileserver"
	"github.com/mintxtinm/packet-monitor/pkg/xdp"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// LoadXdpTracerCommand load xdp tracer command
var LoadXdpTracerCommand = &cli.Command{
	Name:  "load",
	Usage: "Load new XDP dependencies graph exporter",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "ifindex",
			Value: "",
			Usage: "Network interface device name",
		},
		&cli.IntFlag{
			Name:  "interval",
			Value: 30,
			Usage: "Output time interval",
		},

		&cli.StringFlag{
			Name:  "output",
			Value: "",
			Usage: "Output path",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Value: false,
			Usage: "Verbose",
		},
	},
	Action: func(c *cli.Context) (err error) {
		var (
			ifIndexName = c.String("ifindex")
			ins         *xdp.Instance
			g           errgroup.Group
		)

		if !c.IsSet("ifindex") {
			return fmt.Errorf("ifindex not provided")
		}

		logrus.Info("XDP dependencies graph exporter loaded on: ", ifIndexName)

		g.Go(func() error {
			ins = xdp.New(context.Background(), &xdp.Config{
				IfIndexName: ifIndexName,
			})
			return ins.Load(c)
		})

		// reset graph after every cycle
		g.Go(func() (err error) {
			for {
				if ins == nil {
					continue
				}

				time.Sleep(time.Duration(c.Int("interval")) * time.Second)

				logrus.Info(ins.DepGraph.String())

				if c.IsSet("output") {
					if err = ins.DepGraph.Export(); err != nil {
						break
					}
				}

				ins.DepGraph.Reset()
			}
			return
		})

		if c.IsSet("output") {
			g.Go(func() error {
				fileserver.Init()
				return nil
			})
		}

		if err = g.Wait(); err != nil {
			logrus.Fatal(err)
			return
		}

		return
	},
}

// UnloadXdpTracerCommand unload xdp tracer command
var UnloadXdpTracerCommand = &cli.Command{
	Name:  "unload",
	Usage: "Unload XDP dependencies graph exporter",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "ifindex",
			Value: "",
			Usage: "Network interface device name",
		},
	},
	Action: func(c *cli.Context) (err error) {
		var (
			ifIndexName = c.String("ifindex")
		)

		if !c.IsSet("ifindex") {
			return fmt.Errorf("ifindex not provided")
		}

		// Execute bpf loader
		cmd := exec.Command("./main", "unload", ifIndexName)
		cmd.Dir = "./xdp"
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Error(string(out))
			logrus.Errorf("cmd.Run() failed with %s\n", err)
			os.Exit(-1)
		}

		logrus.Info("XDP dependencies graph exporter unloaded on: ", ifIndexName)

		return
	},
}
