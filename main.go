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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/oleewere/ambarictl/ambari"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Version that will be generated during the build as a constant
var Version string

// GitRevString that will be generated during the build as a constant - represents git revision value
var GitRevString string

func main() {
	app := cli.NewApp()
	app.Name = "ambarictl"
	app.Usage = "CLI tool for handle Ambari clusters"
	app.EnableBashCompletion = true
	app.UsageText = "ambarictl command [command options] [arguments...]"
	if len(Version) > 0 {
		app.Version = Version
	} else {
		app.Version = "0.1.0"
	}

	app.Commands = []cli.Command{}
	initCommand := cli.Command{
		Name:  "init",
		Usage: "Initialize Ambari registry database",
		Action: func(c *cli.Context) error {
			ambari.CreateAmbariRegistryDb()
			fmt.Println("Ambari registry DB has been initialized.")
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
			printTable("AMBARI REGISTRIES:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "ACTIVE"}, tableData, c)
			return nil
		},
	}

	listAgentsCommand := cli.Command{
		Name:  "hosts",
		Usage: "Print all registered Ambari agent hosts",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			hosts := ambariRegistry.ListAgents()
			var tableData [][]string
			for _, host := range hosts {
				tableData = append(tableData, []string{host.PublicHostname, host.IP, host.OSType, host.OSArch, strconv.FormatBool(host.UnlimitedJCE), host.HostState})
			}
			printTable("HOSTS:", []string{"PUBLIC HOSTNAME", "IP", "OS TYPE", "OS ARCH", "UNLIMETED_JCE", "STATE"}, tableData, c)
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
			printTable("SERVICES:", []string{"NAME", "STATE"}, tableData, c)
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
			printTable("COMPONENTS:", []string{"NAME", "STATE"}, tableData, c)
			return nil
		},
	}

	listHostComponentsCommand := cli.Command{
		Name:  "hcomponents",
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
				fmt.Println("Flag '--component' or `--host`with a value is required for 'host-components' action!")
				os.Exit(1)
			}
			components := ambariRegistry.ListHostComponents(param, useHost)
			var tableData [][]string
			for _, hostComponent := range components {
				tableData = append(tableData, []string{hostComponent.HostComponentName, hostComponent.HostComponntHost, hostComponent.HostComponentState})
			}
			printTable("HOST COMPONENTS: "+param, []string{"NAME", "HOST", "STATE"}, tableData, c)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "component", Usage: "Component filter for host components"},
			cli.StringFlag{Name: "host", Usage: "Host name filter for host components"},
		},
	}

	createCommand := cli.Command{
		Name:  "create",
		Usage: "Register new Ambari registry entry",
		Action: func(c *cli.Context) error {
			name := ambari.GetStringFlag(c.String("name"), "", "Enter ambari registry name")
			ambariEntryId := ambari.GetAmbariEntryId(name)
			if len(ambariEntryId) > 0 {
				fmt.Println("Ambari registry entry already exists with id " + name)
				os.Exit(1)
			}
			host := ambari.GetStringFlag(c.String("host"), "", "Enter ambari host name")
			portStr := ambari.GetStringFlag(c.String("port"), "8080", "Enter ambari port")
			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			protocol := strings.ToLower(ambari.GetStringFlag(c.String("protocol"), "http", "Enter ambari protocol"))
			if protocol != "http" && protocol != "https" {
				fmt.Println("Use 'http' or 'https' value for protocol option")
				os.Exit(1)
			}
			username := strings.ToLower(ambari.GetStringFlag(c.String("username"), "admin", "Enter ambari user"))
			password := ambari.GetPassword(c.String("password"), "Enter ambari user password")
			cluster := strings.ToLower(ambari.GetStringFlag(c.String("cluster"), "", "Enter ambari cluster"))

			ambari.DeactiveAllAmbariRegistry()
			ambari.RegisterNewAmbariEntry(name, host, port, protocol,
				username, password, cluster)
			fmt.Println("New Ambari registry entry created: " + name)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "Name of the Ambari registry entry"},
			cli.StringFlag{Name: "host", Usage: "Hostname of the Ambari Server"},
			cli.StringFlag{Name: "port", Usage: "Port for Ambari Server"},
			cli.StringFlag{Name: "protocol", Usage: "Protocol for Ambar REST API: http/https"},
			cli.StringFlag{Name: "username", Usage: "User name for Ambari server"},
			cli.StringFlag{Name: "password", Usage: "Password for Ambari user"},
			cli.StringFlag{Name: "cluster", Usage: "Cluster name"},
		},
	}

	deleteCommand := cli.Command{
		Name:  "delete",
		Usage: "De-register an existing Ambari registry entry",
		Action: func(c *cli.Context) error {
			if len(c.Args()) == 0 {
				fmt.Println("Provide a registry name argument for use command. e.g.: delete vagrant")
				os.Exit(1)
			}
			name := c.Args().First()
			ambariEntryId := ambari.GetAmbariEntryId(name)
			if len(ambariEntryId) == 0 {
				fmt.Println("Ambari registry entry does not exist with id " + name)
				os.Exit(1)
			}
			ambari.DeRegisterAmbariEntry(name)
			fmt.Println("Ambari registry de-registered with id: " + name)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "name of the Ambari registry entry"},
		},
	}

	useCommand := cli.Command{
		Name:  "use",
		Usage: "Use selected Ambari registry",
		Action: func(c *cli.Context) error {
			if len(c.Args()) == 0 {
				fmt.Println("Provide a registry name argument for use command. e.g.: use vagrant")
				os.Exit(1)
			}
			name := c.Args().First()
			ambariEntryId := ambari.GetAmbariEntryId(name)
			if len(ambariEntryId) == 0 {
				fmt.Println("Ambari registry entry does not exist with id " + name)
				os.Exit(1)
			}
			ambari.DeactiveAllAmbariRegistry()
			ambari.ActiveAmbariRegistry(name)
			fmt.Println("Ambari registry selected with id: " + name)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "name of the Ambari registry entry"},
		},
	}

	clearCommand := cli.Command{
		Name:  "clear",
		Usage: "Drop all Ambari registry records",
		Action: func(c *cli.Context) error {
			ambari.DropAmbariRegistryRecords()
			fmt.Println("Ambari registry entries dropped.")
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
			printTable("ACTIVE AMBARI REGISTRY:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "ACTIVE"}, tableData, c)
			return nil
		},
	}

	configsCommand := cli.Command{
		Name:  "configs",
		Usage: "Operations with Ambari service configurations",
		Subcommands: []cli.Command{
			{
				Name:  "versions",
				Usage: "Print all service config types with versions",
				Action: func(c *cli.Context) error {
					ambariRegistry := ambari.GetActiveAmbari()
					configs := ambariRegistry.ListServiceConfigVersions()
					var tableData [][]string
					for _, config := range configs {
						tableData = append(tableData, []string{config.ServiceConfigType, strconv.FormatFloat(config.ServiceConfigVersion, 'f', -1, 64), config.ServiceConfigTag})
					}
					printTable("SERVICE_CONFIGS:", []string{"TYPE", "VERSION", "TAG"}, tableData, c)
					return nil
				},
			},
			{
				Name:  "export",
				Usage: "Export cluster configuration to a blueprint json",
				Action: func(c *cli.Context) error {
					ambariRegistry := ambari.GetActiveAmbari()
					var blueprint []byte
					if c.Bool("minimal") {
						clusterInfo := ambariRegistry.GetClusterInfo()
						if len(clusterInfo.ClusterVersion) > 0 {
							splittedString := strings.Split(clusterInfo.ClusterVersion, "-")
							stackName := splittedString[0]
							stackVersion := splittedString[1]
							stackDefaults := ambariRegistry.GetStackDefaultConfigs(stackName, stackVersion)
							largeBlueprint := ambariRegistry.ExportBlueprintAsMap()
							blueprint = ambariRegistry.GetMinimalBlueprint(largeBlueprint, stackDefaults)
							if len(c.String("file")) > 0 {
								err := ioutil.WriteFile(c.String("file"), formatJson(blueprint).Bytes(), 0644)
								if err != nil {
									fmt.Println(err)
									os.Exit(1)
								}
								return nil
							}
						} else {
							fmt.Println("Cannot find a cluster with a name and version for Ambari servrer")
							os.Exit(1)
						}
					} else {
						blueprint = ambariRegistry.ExportBlueprint()
						if len(c.String("file")) > 0 {
							err := ioutil.WriteFile(c.String("file"), blueprint, 0644)
							if err != nil {
								fmt.Println(err)
								os.Exit(1)
							}
							return nil
						}
					}
					printJson(blueprint)
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{Name: "file, f", Usage: "File output for the generated JSON"},
					cli.BoolFlag{Name: "minimal, m", Usage: "Use minimal configuration"},
				},
			},
		},
	}

	clusterCommand := cli.Command{
		Name:  "cluster",
		Usage: "Print Ambari managed cluster details",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			clusterInfo := ambariRegistry.GetClusterInfo()
			var tableData [][]string
			if len(ambariRegistry.Name) > 0 {
				tableData = append(tableData, []string{clusterInfo.ClusterName, clusterInfo.ClusterVersion, clusterInfo.ClusterSecurityType, strconv.FormatFloat(clusterInfo.ClusterTotalHosts, 'f', -1, 64)})
			}
			printTable("CLUSTER INFO:", []string{"Name", "VERSION", "SECURITY", "TOTAL HOSTS"}, tableData, c)
			return nil
		},
	}

	app.Commands = append(app.Commands, initCommand)
	app.Commands = append(app.Commands, createCommand)
	app.Commands = append(app.Commands, deleteCommand)
	app.Commands = append(app.Commands, useCommand)
	app.Commands = append(app.Commands, showCommand)
	app.Commands = append(app.Commands, listCommand)
	app.Commands = append(app.Commands, listAgentsCommand)
	app.Commands = append(app.Commands, listServicesCommand)
	app.Commands = append(app.Commands, listComponentsCommand)
	app.Commands = append(app.Commands, listHostComponentsCommand)
	app.Commands = append(app.Commands, configsCommand)
	app.Commands = append(app.Commands, clusterCommand)
	app.Commands = append(app.Commands, clearCommand)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printTable(title string, headers []string, data [][]string, c *cli.Context) {
	fmt.Println(title)
	if len(data) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(headers)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	} else {
		for i := 1; i <= len(title); i++ {
			fmt.Print("-")
		}
		fmt.Println()
		fmt.Println("NO ENTRIES FOUND!")
	}
}

func printJson(b []byte) {
	fmt.Println(formatJson(b).String())
}

func formatJson(b []byte) *bytes.Buffer {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &out
}
