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
	"os/user"
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
	if len(GitRevString) > 0 {
		app.Version = app.Version + fmt.Sprintf(" (git short hash: %v)", GitRevString)
	}

	app.Commands = []cli.Command{}
	initCommand := cli.Command{
		Name:  "init",
		Usage: "Initialize Ambari server database",
		Action: func(c *cli.Context) error {
			ambari.CreateAmbariRegistryDb()
			fmt.Println("Ambari registry DB has been initialized.")
			return nil
		},
	}

	listCommand := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "Print all registered Ambari servers",
		Action: func(c *cli.Context) error {
			ambariServerEntries := ambari.ListAmbariRegistryEntries()
			var tableData [][]string
			for _, ambariServer := range ambariServerEntries {
				activeValue := "false"
				if ambariServer.Active {
					activeValue = "true"
				}
				tableData = append(tableData, []string{ambariServer.Name, ambariServer.Hostname, strconv.Itoa(ambariServer.Port), ambariServer.Protocol,
					ambariServer.Username, "********", ambariServer.Cluster, ambariServer.ConnectionProfile, activeValue})
			}
			printTable("AMBARI REGISTRIES:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "PROFILE", "ACTIVE"}, tableData, c)
			return nil
		},
	}

	profileCommand := cli.Command{
		Name:  "profiles",
		Usage: "Connection profiles related commands",
		Subcommands: []cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create new connection profile",
				Action: func(c *cli.Context) error {
					name := ambari.GetStringFlag(c.String("name"), "", "Enter connection profile name")
					connProfileId := ambari.GetConnectionProfileEntryId(name)
					if len(connProfileId) > 0 {
						fmt.Println("Connection profile entry already exists with id " + name)
						os.Exit(1)
					}
					keyPath := ambari.GetStringFlag(c.String("key_path"), "", "Enter ssh key path")
					usr, err := user.Current()
					if err != nil {
						panic(err)
					}
					home := usr.HomeDir
					keyPath = strings.Replace(keyPath, "~", home, -1)
					if len(keyPath) > 0 {
						if _, err := os.Stat(keyPath); err != nil {
							if os.IsNotExist(err) {
								fmt.Println(err)
								os.Exit(1)
							}
						}
					}
					portStr := ambari.GetStringFlag(c.String("port"), "22", "Enter ssh port")
					port, err := strconv.Atoi(portStr)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					userName := ambari.GetStringFlag(c.String("username"), "root", "Enter ssh username")
					hostJumpStr := ambari.GetStringFlag(c.String("host_jump"), "n", "Use host jump?")
					hostJump := ambari.EvaluateBoolValueFromString(hostJumpStr)
					proxyAddress := ""
					if hostJump {
						proxyAddress = ambari.GetStringFlag(c.String("proxy_address"), "none", "Set a proxy address?")
						if proxyAddress == "none" {
							proxyAddress = ""
						}
					}
					ambari.RegisterNewConnectionProfile(name, keyPath, port, userName, hostJump, proxyAddress)
					fmt.Println("New connection profile entry has been created: " + name)
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{Name: "name", Usage: "Name of the Ambari server entry"},
					cli.StringFlag{Name: "key_path", Usage: "Hostname of the Ambari server"},
					cli.StringFlag{Name: "port", Usage: "Port for AmbarisServer"},
					cli.StringFlag{Name: "username", Usage: "Protocol for Ambar REST API: http/https"},
					cli.StringFlag{Name: "host_jump", Usage: "User name for Ambari server"},
					cli.StringFlag{Name: "proxy_address", Usage: "Password for Ambari user"},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "Print all connection profile entries",
				Action: func(c *cli.Context) error {
					connectionProfiles := ambari.ListConnectionProfileEntries()
					var tableData [][]string
					for _, profile := range connectionProfiles {
						hostJump := "false"
						if profile.HostJump {
							hostJump = "true"
						}
						tableData = append(tableData, []string{profile.Name, profile.KeyPath, strconv.Itoa(profile.Port), profile.Username, hostJump, profile.ProxyAddress})
					}
					printTable("CONNECTION PROFILES:", []string{"NAME", "KEY", "PORT", "USERNAME", "HOST JUMP", "PROXY ADDRESS"}, tableData, c)
					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete a connection profile entry by id",
				Action: func(c *cli.Context) error {
					if len(c.Args()) == 0 {
						fmt.Println("Provide a profile name argument for use command. e.g.: delete vagrant")
						os.Exit(1)
					}
					name := c.Args().First()
					profileEntryId := ambari.GetConnectionProfileEntryId(name)
					if len(profileEntryId) == 0 {
						fmt.Println("Connection profile entry does not exist with id " + name)
						os.Exit(1)
					}
					ambari.DeRegisterConnectionProfile(profileEntryId)
					msg := fmt.Sprintf("Connection profile '%s' has been deleted successfully", profileEntryId)
					fmt.Println(msg)
					return nil
				},
			},
			{
				Name:    "clear",
				Aliases: []string{"cl"},
				Usage:   "Delete all connection profile entries",
				Action: func(c *cli.Context) error {
					ambari.DropConnectionProfileRecords()
					fmt.Println("All connection profile records has been dropped")
					return nil
				},
			},
		},
	}

	attachCommand := cli.Command{
		Name:  "attach",
		Usage: "Attach a profile to an ambari server entry",
		Action: func(c *cli.Context) error {
			args := c.Args()
			if len(args) == 0 {
				fmt.Println("Provide at least 1 argument (<profile>), or 2 (<profile> and <ambariEntry>)")
				os.Exit(1)
			}
			profileId := args.Get(0)
			var ambariRegistry ambari.AmbariRegistry
			if len(args) == 1 {
				ambariRegistry = ambari.GetActiveAmbari()
				if len(ambariRegistry.Name) == 0 {
					fmt.Println("No active ambari selected")
					os.Exit(1)
				}
			} else {
				ambariRegistryId := args.Get(1)
				ambari.GetAmbariById(ambariRegistryId)
				if len(ambariRegistry.Name) == 0 {
					fmt.Println("Cannot find specific ambari server entry")
					os.Exit(1)
				}
			}
			profile := ambari.GetConnectionProfileById(profileId)
			if len(profile.Name) == 0 {
				fmt.Println("Cannot find specific connection profile entry")
				os.Exit(1)
			}

			ambari.SetProfileIdForAmbariEntry(ambariRegistry.Name, profile.Name)
			msg := fmt.Sprintf("Attach profile '%s' to '%s'", profile.Name, ambariRegistry.Name)
			fmt.Println(msg)
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
			printTable("HOSTS:", []string{"PUBLIC HOSTNAME", "IP", "OS TYPE", "OS ARCH", "UNLIMITED_JCE", "STATE"}, tableData, c)
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
				tableData = append(tableData, []string{component.ComponentName, component.ServiceName, component.ComponentState})
			}
			printTable("COMPONENTS:", []string{"NAME", "SERVICE", "STATE"}, tableData, c)
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
		Usage: "Register new Ambari server entry",
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
			cluster := ambari.GetStringFlag(c.String("cluster"), "", "Enter ambari cluster")

			ambari.DeactiveAllAmbariRegistry()
			ambari.RegisterNewAmbariEntry(name, host, port, protocol,
				username, password, cluster)
			fmt.Println("New Ambari server entry has been created: " + name)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "Name of the Ambari server entry"},
			cli.StringFlag{Name: "host", Usage: "Hostname of the Ambari server"},
			cli.StringFlag{Name: "port", Usage: "Port for AmbarisServer"},
			cli.StringFlag{Name: "protocol", Usage: "Protocol for Ambar REST API: http/https"},
			cli.StringFlag{Name: "username", Usage: "User name for Ambari server"},
			cli.StringFlag{Name: "password", Usage: "Password for Ambari user"},
			cli.StringFlag{Name: "cluster", Usage: "Cluster name"},
		},
	}

	deleteCommand := cli.Command{
		Name:  "delete",
		Usage: "De-register an existing Ambari server entry",
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
		Usage: "Use selected Ambari server",
		Action: func(c *cli.Context) error {
			if len(c.Args()) == 0 {
				fmt.Println("Provide a server entry name argument for use command. e.g.: use vagrant")
				os.Exit(1)
			}
			name := c.Args().First()
			ambariEntryId := ambari.GetAmbariEntryId(name)
			if len(ambariEntryId) == 0 {
				fmt.Println("Ambari server entry does not exist with id " + name)
				os.Exit(1)
			}
			ambari.DeactiveAllAmbariRegistry()
			ambari.ActiveAmbariRegistry(name)
			fmt.Println("Ambari server entry selected with id: " + name)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "name of the Ambari registry entry"},
		},
	}

	clearCommand := cli.Command{
		Name:  "clear",
		Usage: "Drop all Ambari server records",
		Action: func(c *cli.Context) error {
			ambari.DropAmbariRegistryRecords()
			fmt.Println("Ambari server entries dropped.")
			return nil
		},
	}

	showCommand := cli.Command{
		Name:  "show",
		Usage: "Show active Ambari server details",
		Action: func(c *cli.Context) error {
			ambariRegistry := ambari.GetActiveAmbari()
			var tableData [][]string
			if len(ambariRegistry.Name) > 0 {
				tableData = append(tableData, []string{ambariRegistry.Name, ambariRegistry.Hostname, strconv.Itoa(ambariRegistry.Port), ambariRegistry.Protocol,
					ambariRegistry.Username, "********", ambariRegistry.Cluster, ambariRegistry.ConnectionProfile, "true"})
			}
			printTable("ACTIVE AMBARI REGISTRY:", []string{"Name", "HOSTNAME", "PORT", "PROTOCOL", "USER", "PASSWORD", "CLUSTER", "PROFILE", "ACTIVE"}, tableData, c)
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
				Name:  "update",
				Usage: "Update config value for a specific config key of a config type",
				Action: func(c *cli.Context) error {
					ambariRegistry := ambari.GetActiveAmbari()
					if len(c.String("config-type")) == 0 {
						fmt.Println("Parameter '--config-type' is required")
						os.Exit(1)
					}
					if len(c.String("config-key")) == 0 {
						fmt.Println("Parameter '--config-key' is required")
						os.Exit(1)
					}
					if len(c.String("config-value")) == 0 {
						fmt.Println("Parameter '--config-value' is required")
						os.Exit(1)
					}
					ambariRegistry.SetConfig(c.String("type"), c.String("key"), c.String("value"))
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{Name: "type, t", Usage: "Configuration type"},
					cli.StringFlag{Name: "key, k", Usage: "Configuration key"},
					cli.StringFlag{Name: "value, v", Usage: "Configuration value"},
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

	runCommand := cli.Command{
		Name:  "run",
		Usage: "Execute commands on all (or specific) hosts",
		Action: func(c *cli.Context) error {
			ambariServer := ambari.GetActiveAmbari()
			args := c.Args()
			command := ""
			for _, arg := range args {
				command += arg
			}
			filter := ambari.CreateFilter(strings.ToUpper(c.String("services")),
				strings.ToUpper(c.String("components")), c.String("hosts"), c.Bool("server"))
			hosts := ambariServer.GetFilteredHosts(filter)
			ambariServer.RunRemoteHostCommand(command, hosts)
			return nil
		},
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "server", Usage: "Filter on ambari-server"},
			cli.StringFlag{Name: "services, s", Usage: "Filter on services (comma separated)"},
			cli.StringFlag{Name: "components, c", Usage: "Filter on components (comma separated)"},
			cli.StringFlag{Name: "hosts", Usage: "Filter on hosts (comma separated)"},
		},
	}

	commandCommand := cli.Command{
		Name:  "command",
		Usage: "Execute ambari commands on Ambari server (START/STOP/RESTART)",
		Action: func(c *cli.Context) error {
			ambariServer := ambari.GetActiveAmbari()
			args := c.Args()
			command := ""
			for _, arg := range args {
				command += arg
			}
			if len(c.String("services")) == 0 && len(c.String("components")) == 0 {
				fmt.Println("It is required to provide --components (-c) or --services (-s) flag")
				os.Exit(1)
			}
			filter := ambari.CreateFilter(strings.ToUpper(c.String("services")),
				strings.ToUpper(c.String("components")), "", false)
			ambariServer.RunAmbariServiceCommand(command, filter, len(filter.Services) > 0, len(filter.Components) > 0)
			if len(c.String("components")) > 0 {
				fmt.Println(fmt.Sprintf("Command %s has been sent to %s (components)", command, c.String("components")))
			} else if len(c.String("services")) > 0 {
				fmt.Println(fmt.Sprintf("Command %s has been sent to %s (services)", command, c.String("services")))
			}

			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "services, s", Usage: "Filter on services (comma separated)"},
			cli.StringFlag{Name: "components, c", Usage: "Filter on components (comma separated)"},
		},
	}

	playbookCommand := cli.Command{
		Name:  "playbook",
		Usage: "Execute a list of commands defined in playbook file(s)",
		Action: func(c *cli.Context) error {
			ambariServer := ambari.GetActiveAmbari()
			if len(c.String("file")) == 0 {
				fmt.Println("Provide --file parameter")
				os.Exit(1)
			}
			playbook := ambari.LoadPlaybookFile(c.String("file"))
			ambariServer.ExecutePlaybook(playbook)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "file, f", Usage: "Playbook file"},
		},
	}

	logsCommand := cli.Command{
		Name:  "logs",
		Usage: "Download logs from Ambari agents",
		Action: func(c *cli.Context) error {
			ambariServer := ambari.GetActiveAmbari()
			if len(c.String("destination")) == 0 {
				fmt.Println("Provide --destination parameter")
				os.Exit(1)
			}
			filter := ambari.CreateFilter(strings.ToUpper(c.String("services")),
				strings.ToUpper(c.String("components")), c.String("hosts"), c.Bool("server"))
			ambariServer.DownloadLogs(c.String("destination"), filter)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "destination, d", Usage: "Download destination"},
			cli.BoolFlag{Name: "server", Usage: "Download server logs flag"},
			cli.StringFlag{Name: "services, s", Usage: "Filter on services (comma separated)"},
			cli.StringFlag{Name: "components, c", Usage: "Filter on components (comma separated)"},
			cli.StringFlag{Name: "hosts", Usage: "Filter on hosts (comma separated)"},
		},
	}

	app.Commands = append(app.Commands, initCommand)
	app.Commands = append(app.Commands, createCommand)
	app.Commands = append(app.Commands, deleteCommand)
	app.Commands = append(app.Commands, useCommand)
	app.Commands = append(app.Commands, showCommand)
	app.Commands = append(app.Commands, runCommand)
	app.Commands = append(app.Commands, commandCommand)
	app.Commands = append(app.Commands, playbookCommand)
	app.Commands = append(app.Commands, profileCommand)
	app.Commands = append(app.Commands, attachCommand)
	app.Commands = append(app.Commands, listCommand)
	app.Commands = append(app.Commands, listAgentsCommand)
	app.Commands = append(app.Commands, listServicesCommand)
	app.Commands = append(app.Commands, listComponentsCommand)
	app.Commands = append(app.Commands, listHostComponentsCommand)
	app.Commands = append(app.Commands, configsCommand)
	app.Commands = append(app.Commands, clusterCommand)
	app.Commands = append(app.Commands, logsCommand)
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
