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

package ambari

import (
	"fmt"
)

var logDirMap = map[string]map[string]map[string]string{
	"ZOOKEEPER": {
		"ZOOKEEPER_SERVER": {
			"log_dir_config_type":     "zookeeper-env",
			"log_dir_config_property": "zk_log_dir",
		},
		"ZOOKEEPER_CLIENT": {
			"log_dir_config_type":     "zookeeper-env",
			"log_dir_config_property": "zk_log_dir",
		},
	},
	"AMBARI_INFRA_SOLR": {
		"INFRA_SOLR": {
			"log_dir_config_type":     "infra-solr-env",
			"log_dir_config_property": "infra_solr_log_dir",
		},
		"INFRA_SOLR_CLIENT": {
			"log_dir_config_type":     "infra-solr-client-log4j",
			"log_dir_config_property": "infra_solr_client_log_dir",
		},
	},
	"LOGSEARCH": {
		"LOGSEARCH_SERVER": {
			"log_dir_config_type":     "logsearch-env",
			"log_dir_config_property": "logsearch_log_dir",
		},
		"LOGSEARCH_LOGFEEDER": {
			"log_dir_config_type":     "logfeeder-env",
			"log_dir_config_property": "logfeeder_log_dir",
		},
	},
}

// DownloadLogs download specific logs that can be filtered by hosts, components or service (by default, it downloads agent logs)
func (a AmbariRegistry) DownloadLogs(dest string, filter Filter) {
	componentLogDirMap := getComponentLogDirMap(a, filter)
	if filter.Server {
		serverHosts := a.GetFilteredHosts(filter)
		fmt.Println(serverHosts)
	} else {
		if len(componentLogDirMap) > 0 {
			if len(filter.Services) > 0 {
				for _, service := range filter.Services {
					hostComponents := a.ListHostComponentsByService(service)
					componentMap := make(map[string]bool)
					for _, hostComponent := range hostComponents {
						componentMap[hostComponent.HostComponentName] = true
					}
					for component, _ := range componentMap {
						componentFilter := Filter{Hosts: filter.Hosts, Components: []string{component}}
						hosts := a.GetFilteredHosts(componentFilter)
						fmt.Println(component)
						fmt.Println(hosts)
					}
				}
			}
			if len(filter.Components) > 0 {
				for _, component := range filter.Components {
					componentFilter := Filter{Hosts: filter.Hosts, Components: []string{component}}
					hosts := a.GetFilteredHosts(componentFilter)
					fmt.Println(component)
					fmt.Println(hosts)
				}
			}
		} else {
			hosts := a.GetFilteredHosts(filter)
			fmt.Println(hosts)
		}
	}
	return
}

func getComponentLogDirMap(ambariRegistry AmbariRegistry, filter Filter) map[string]string {
	componentLogDirMap := map[string]string{}
	if len(filter.Services) > 0 || len(filter.Components) > 0 {
		blueprint := ambariRegistry.ExportBlueprintAsMap()
		services := make([]string, len(logDirMap))
		if len(filter.Services) > 0 {
			services = filter.Services
		} else {
			for service := range logDirMap {
				services = append(services, service)
			}
		}
		for _, service := range services {
			if components, ok := logDirMap[service]; ok {
				filteredComponents := []string{}
				if len(filter.Components) > 0 {
					filteredComponents = filter.Components
				}
				for componentKey, component := range components {
					var logConfigType = ""
					var logConfigProperty = ""
					var logDirDefault = ""
					if logConfigTypeVal, ok := component["log_dir_config_type"]; ok {
						logConfigType = logConfigTypeVal
					}
					if logConfigPropertyVal, ok := component["log_dir_config_property"]; ok {
						logConfigProperty = logConfigPropertyVal
					}
					if logDirDefaultVal, ok := component["log_dir_default"]; ok {
						logDirDefault = logDirDefaultVal
					}
					componentLogDir := ""
					if len(logConfigProperty) > 0 && len(logConfigType) > 0 {
						componentLogDir = GetConfigValue(blueprint, logConfigType, logConfigProperty)
					} else if len(logDirDefault) > 0 {
						componentLogDir = logDirDefault
					}
					if len(componentLogDir) > 0 {
						if len(filteredComponents) > 0 {
							for _, comp := range filteredComponents {
								if comp == componentKey {
									componentLogDirMap[componentKey] = componentLogDir
								}
							}
						} else {
							componentLogDirMap[componentKey] = componentLogDir
						}
					}
				}
			}
		}
	}
	return componentLogDirMap
}
