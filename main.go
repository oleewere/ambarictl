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
	"github.com/olekukonko/tablewriter"
	"strconv"
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
			ambariServerEntries := ambari.ListAmbariRegistryEntries()
			var tableData [][]string
			for _, ambariServer := range ambariServerEntries {
				activeValue := "false"
				if ambariServer.Active == 1 {
					activeValue = "true"
				}
				tableData = append(tableData, []string{ambariServer.Name, ambariServer.Hostname, strconv.Itoa(ambariServer.Port), ambariServer.Protocol,
				ambariServer.Username, "********", ambariServer.Cluster, activeValue})
			}
			printTable("AMBARI REGISTRIES:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "ACTIVE"}, tableData)
			return nil
		},
	}

	listAgentsCommand := cli.Command{
		Name:  "hosts",
		Usage: "Print all registered Ambari registries",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			hosts := ambariRegistry.ListAgents()
			var tableData [][]string
			for _, host := range hosts {
				tableData = append(tableData, []string{host.PublicHostname, host.IP, host.HostState})
			}
			printTable("HOSTS:", []string{"PUBLIC HOSTNAME", "IP", "STATE"}, tableData)
			return nil
		},
	}

	listServicesCommand := cli.Command{
		Name:  "services",
		Usage: "Print all installed Ambari services",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			services := ambariRegistry.ListServices()
			var tableData [][]string
			for _, service := range services {
				tableData = append(tableData, []string{service.ServiceName, service.ServiceState})
			}
			printTable("SERVICES:", []string{"NAME", "STATE"}, tableData)
			return nil
		},
	}

	listComponentsCommand := cli.Command{
		Name:  "components",
		Usage: "Print all installed Ambari components",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			components := ambariRegistry.ListComponents()
			var tableData [][]string
			for _, component := range components {
				tableData = append(tableData, []string{component.ComponentName, component.ComponentState})
			}
			printTable("COMPONENTS:", []string{"NAME", "STATE"}, tableData)
			return nil
		},
	}

	listHostComponentsCommand := cli.Command{
		Name:  "host-components",
		Usage: "Print all installed Ambari host components by component name",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			var param string
			useHost := false
			if len(c.String("component")) > 0 {
				param = c.String("component")
			} else if len(c.String("host")) > 0 {
				param = c.String("host")
				useHost = true
			} else {
				fmt.Println("Flag '--component' or `--`with a value is required for 'host-components' action!")
				os.Exit(1)
			}
			components := ambariRegistry.ListHostComponents(param, useHost)
			var tableData [][]string
			for _, hostComponent := range components {
				tableData = append(tableData, []string{hostComponent.HostComponentName, hostComponent.HostComponntHost, hostComponent.HostComponentState})
			}
			printTable("HOST COMPONENTS: " + param, []string{"NAME", "HOST", "STATE"}, tableData)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "component", Usage: "Component filter for host components"},
			cli.StringFlag{Name: "host", Usage: "Host name filter for host components"},
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
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "Name of the Ambari registry entry"},
			cli.StringFlag{Name: "host", Usage: "Hostname of the Ambari Server"},
			cli.IntFlag{Name: "port", Usage: "Port for Ambari Server"},
			cli.BoolFlag{Name: "ssl", Usage: "Enabled TLS/SSL for Ambari"},
			cli.StringFlag{Name: "username", Usage: "Ambari user"},
			cli.StringFlag{Name: "password", Usage: "Password for Ambari user"},
			cli.StringFlag{Name: "cluster", Usage: "Cluster name"},
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
			var tableData [][]string
			if len(ambariRegistry.Name) > 0 {
				tableData = append(tableData, []string{ambariRegistry.Name, ambariRegistry.Hostname, strconv.Itoa(ambariRegistry.Port), ambariRegistry.Protocol,
					ambariRegistry.Username, "********", ambariRegistry.Cluster, "true"})
			}
			printTable("ACTIVE AMBARI REGISTRY:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "ACTIVE"}, tableData)
			return nil
		},
	}

	app.Commands = append(app.Commands, initCommand)
	app.Commands = append(app.Commands, listCommand)
	app.Commands = append(app.Commands, listAgentsCommand)
	app.Commands = append(app.Commands, listServicesCommand)
	app.Commands = append(app.Commands, listComponentsCommand)
	app.Commands = append(app.Commands, listHostComponentsCommand)
	app.Commands = append(app.Commands, showCommand)
	app.Commands = append(app.Commands, registerCommand)
	app.Commands = append(app.Commands, clearCommand)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printTable(title string, headers []string, data [][]string) {
	fmt.Println(title)
	if len(data) > 0 {table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(headers)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	} else {
		for i:= 1; i <= len(title); i++ {
			fmt.Print("-")
		}
		fmt.Println()
		fmt.Println("NO ENTRIES FOUND!")
	}
}
