// Copyright 2018 Oliver Szabo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/oleewere/ambari-manager/ambari"
	"github.com/urfave/cli"
	"os"
)

// Version that will be generated during the build as a constant
var Version string

// GitRevString that will be generated during the build as a constant - represents git revision value
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
	initCommand := cli.Command{
		Name:  "init",
		Usage: "Initialize Ambari context (db)",
		Action: func(c *cli.Context) error {
			ambari.CreateAmbariRegistryDb()
			return nil
		},
	}
	listCommand := cli.Command{
		Name:  "list",
		Usage: "Print all registered Ambari registries",
		Action: func(c *cli.Context) error {
			fmt.Println("Ambari registries:")
			fmt.Println("------------------")
			ambari.ListAmbariRegistryEntries()
			return nil
		},
	}

	listAgentsCommand := cli.Command{
		Name:  "hosts",
		Usage: "Print all registered Ambari registries",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			hosts := ambariRegistry.ListAgents()
			fmt.Println("Registered hosts:")
			fmt.Println("-----------------")
			for _, host := range hosts {
				hostEntry := fmt.Sprintf("%s (ip: %s) - state: %s", host.PublicHostname, host.IP, host.HostState)
				fmt.Println(hostEntry)
			}
			return nil
		},
	}

	listServicesCommand := cli.Command{
		Name:  "services",
		Usage: "Print all installed Ambari services",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			services := ambariRegistry.ListServices()
			fmt.Println("Installed services:")
			fmt.Println("-------------------")
			for _, service := range services {
				serviceEntry := fmt.Sprintf("%s (state: %s)", service.ServiceName, service.ServiceState)
				fmt.Println(serviceEntry)
			}
			return nil
		},
	}

	registerCommand := cli.Command{
		Name:  "register",
		Usage: "Register new Ambari entry",
		Action: func(c *cli.Context) error {
			ambari.RegisterNewAmbariEntry("vagrant", "c7401.ambari.apache.org", 8080, "http",
				"admin", "admin", "cl1")
			return nil
		},
	}

	clearCommand := cli.Command{
		Name:  "clear",
		Usage: "Drop all Ambari registry records",
		Action: func(c *cli.Context) error {
			ambari.DropAmbariRegistryRecords()
			return nil
		},
	}

	showCommand := cli.Command{
		Name:  "show",
		Usage: "Show active Ambari registry details",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			fmt.Println("Active Ambari registry:")
			fmt.Println("-----------------------")
			ambariRegistry.ShowDetails()
			return nil
		},
	}

	app.Commands = append(app.Commands, initCommand)
	app.Commands = append(app.Commands, listCommand)
	app.Commands = append(app.Commands, listAgentsCommand)
	app.Commands = append(app.Commands, listServicesCommand)
	app.Commands = append(app.Commands, showCommand)
	app.Commands = append(app.Commands, registerCommand)
	app.Commands = append(app.Commands, clearCommand)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
