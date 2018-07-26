package main

import (
	"github.com/urfave/cli"
	"os"
	"github.com/oleewere/ambari-manager/ambari"
	"fmt"
)

var Version string
var GitRevString string

func main() {
	app := cli.NewApp()
	app.Name = "ambari-manager"
	app.Usage = "CLI tool for handle Ambari clusters"
	app.EnableBashCompletion = true
	app.UsageText = "ambari-manager command [command options] [arguments...]"
	if len(Version) > 0 {
		app.Version = Version
	} else {
		app.Version = "0.1.0"
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "verbose"},
	}

	app.Commands = []cli.Command{}
	init_command := cli.Command{
		Name: "init",
		Usage: "Initialize Ambari context (db)",
		Action: func(c *cli.Context) error {
			ambari.CreateAmbariRegistryDb()
			return nil
		},
	}
	list_command := cli.Command{
		Name: "list",
		Usage: "Print all registered Ambari registries",
		Action: func(c *cli.Context) error {
			ambari.ListAmbariRegistryEntries()
			return nil
		},
	}

	list_agents_command := cli.Command{
		Name: "hosts",
		Usage: "Print all registered Ambari registries",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			ambariRegistry.ListAgents()
			return nil
		},
	}

	register_command := cli.Command{
		Name: "register",
		Usage: "Register new Ambari entry",
		Action: func(c *cli.Context) error {
			ambari.RegisterNewAmbariEntry("vagrant", "c7401.ambari.apache.org", 8080, "http",
				"admin", "admin", "cl1")
			return nil
		},
	}

	clear_command := cli.Command{
		Name: "clear",
		Usage: "Drop all Ambari registry records",
		Action: func(c *cli.Context) error {
			ambari.DropAmbariRegistryRecords()
			return nil
		},
	}

	show_command := cli.Command{
		Name: "show",
		Usage: "Show active Ambari registry details",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			fmt.Println("Active Ambari registry:")
			fmt.Println("-----------------------")
			ambariRegistry.ShowDetails()
			return nil
		},
	}

	app.Commands = append(app.Commands, init_command)
	app.Commands = append(app.Commands, list_command)
	app.Commands = append(app.Commands, list_agents_command)
	app.Commands = append(app.Commands, show_command)
	app.Commands = append(app.Commands, register_command)
	app.Commands = append(app.Commands, clear_command)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}