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
	"os"
	"path"
	"strings"
	"time"
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
	"ACCUMULO": {
		"ACCUMULO_MASTER": {
			"log_dir_config_type":     "accumulo-env",
			"log_dir_config_property": "accumulo_log_dir",
		},
	},
	"AMBARI_METRICS": {
		"METRICS_COLLECTOR": {
			"log_dir_config_type":     "ams-env",
			"log_dir_config_property": "metrics_collector_log_dir",
		},
		"METRICS_MONITOR": {
			"log_dir_config_type":     "ams-env",
			"log_dir_config_property": "metrics_monitor_log_dir",
		},
		"METRICS_GRAFANA": {
			"log_dir_config_type":     "ams-grafana-env",
			"log_dir_config_property": "metrics_grafana_log_dir",
		},
	},
	"ATLAS": {
		"ATLAS_MASTER": {
			"log_dir_config_type":     "atlas-env",
			"log_dir_config_property": "metadata_log_dir",
		},
	},
	"DRUID": {
		"DRUID_BROKER": {
			"log_dir_config_type":     "druid-env",
			"log_dir_config_property": "druid_log_dir",
		},
	},
	"HBASE": {
		"HBASE_MASTER": {
			"log_dir_config_type":     "hbase-env",
			"log_dir_config_property": "hbase_log_dir",
		},
		"HBASE_REGIONSERVER": {
			"log_dir_config_type":     "hbase-env",
			"log_dir_config_property": "hbase_log_dir",
		},
	},
	"HDFS": {
		"NAMENODE": {
			"log_dir_config_type":     "hadoop-env",
			"log_dir_config_property": "hdfs_log_dir_prefix",
		},
		"DATANODE": {
			"log_dir_config_type":     "hadoop-env",
			"log_dir_config_property": "hdfs_log_dir_prefix",
		},
	},
	"HIVE": {
		"HIVE_METASTORE": {
			"log_dir_config_type":     "hive-env",
			"log_dir_config_property": "hive_log_dir",
		},
		"HIVE_SERVER": {
			"log_dir_config_type":     "hive-env",
			"log_dir_config_property": "hive_log_dir",
		},
		"HIVE_SERVER_INTERACTIVE": {
			"log_dir_config_type":     "hive-env",
			"log_dir_config_property": "hive_log_dir",
		},
	},
	"KAFKA": {
		"KAFKA_BROKER": {
			"log_dir_config_type":     "kafka-env",
			"log_dir_config_property": "kafka_log_dir",
		},
	},
	"OOZIE": {
		"OOZIE_SERVER": {
			"log_dir_config_type":     "oozie-env",
			"log_dir_config_property": "oozie_log_dir",
		},
	},
	"RANGER": {
		"RANGER_ADMIN": {
			"log_dir_config_type":     "ranger-env",
			"log_dir_config_property": "ranger_admin_log_dir",
		},
		"RANGER_USERSYNC": {
			"log_dir_config_type":     "ranger-env",
			"log_dir_config_property": "ranger_usersync_log_dir",
		},
	},
	"RANGER_KMS": {
		"RANGER_KMS_SERVER": {
			"log_dir_config_type":     "kms-env",
			"log_dir_config_property": "kms_log_dir",
		},
	},
	"SPARK2": {
		"SPARK2_JOBHISTORYSERVER": {
			"log_dir_config_type":     "spark2-env",
			"log_dir_config_property": "spark_log_dir",
		},
		"SPARK2_THRIFTSERVER": {
			"log_dir_config_type":     "spark2-env",
			"log_dir_config_property": "spark_log_dir",
		},
		"LIVY2_SERVER": {
			"log_dir_config_type":     "livy2-env",
			"log_dir_config_property": "livy2_log_dir",
		},
	},
	"SUPERSET": {
		"SUPERSET": {
			"log_dir_config_type":     "superset-env",
			"log_dir_config_property": "superset_log_dir",
		},
	},
	"STORM": {
		"NIMBUS": {
			"log_dir_config_type":     "storm-env",
			"log_dir_config_property": "storm_log_dir",
		},
		"STORM_UI_SERVER": {
			"log_dir_config_type":     "storm-env",
			"log_dir_config_property": "storm_log_dir",
		},
		"SUPERVISOR": {
			"log_dir_config_type":     "storm-env",
			"log_dir_config_property": "storm_log_dir",
		},
	},
	"MAPREDUCE2": {
		"HISTORYSERVER": {
			"log_dir_config_type":     "mapred-env",
			"log_dir_config_property": "mapred_log_dir_prefix",
		},
	},
	"YARN": {
		"APP_TIMELINE_SERVER": {
			"log_dir_config_type":     "yarn-env",
			"log_dir_config_property": "yarn_log_dir_prefix",
		},
		"RESOURCEMANAGER": {
			"log_dir_config_type":     "yarn-env",
			"log_dir_config_property": "yarn_log_dir_prefix",
		},
	},
	"ZEPPELIN": {
		"ZEPPELIN_MASTER": {
			"log_dir_config_type":     "zeppelin-env",
			"log_dir_config_property": "zeppelin_log_dir",
		},
	},
	"SMARTSENSE": {
		"HST_SERVER": {
			"log_dir_config_type":     "hst-log4j",
			"log_dir_config_property": "hst.log.dir",
		},
		"HST_AGENT": {
			"log_dir_config_type":     "hst-log4j",
			"log_dir_config_property": "hst.log.dir",
		},
	},
	"NIFI": {
		"NIFI_CA": {
			"log_dir_config_type":     "nifi-env",
			"log_dir_config_property": "nifi_node_log_dir",
		},
	},
	"STREAMLINE": {
		"STREAMLINE_SERVER": {
			"log_dir_config_type":     "streamline-env",
			"log_dir_config_property": "streamline_log_dir",
		},
	},
}

// DownloadLogs download specific logs that can be filtered by hosts, components or service (by default, it downloads agent logs)
func (a AmbariRegistry) DownloadLogs(dest string, filter Filter) {
	componentLogDirMap := getComponentLogDirMap(a, filter)
	downloadFolder := createDownloadRootFolder(dest, a.Name)
	if filter.Server {
		serverHosts := a.GetFilteredHosts(filter)
		getLogDirCommand := "cat /etc/ambari-server/conf/log4j.properties | grep ambari.root.dir"
		responses := a.RunRemoteHostCommand(getLogDirCommand, serverHosts, filter.Server)
		ambariLogDir := "/var/log/ambari-server"
		for _, response := range responses {
			splittedResponses := strings.Split(response.StdOut, "\n")
			propertyMap := ConvertStingsToMap(splittedResponses)
			ambariRootDir := propertyMap["ambari.root.dir"]
			ambariLogDirUnformatted := strings.Replace(propertyMap["ambari.log.dir"], "${ambari.root.dir}", ambariRootDir, 1)
			ambariLogDir = strings.Replace(ambariLogDirUnformatted, "//", "/", -1)
		}
		fmt.Println(ambariLogDir)
		componentName := "ambari-server"
		componentDownloadFolder := createDownloadFolder(downloadFolder, componentName)
		a.CopyFolderFromRemote(componentName, ambariLogDir, componentDownloadFolder, serverHosts, filter.Server)
	} else {
		if len(componentLogDirMap) > 0 {
			if len(filter.Services) > 0 {
				for _, service := range filter.Services {
					hostComponents := a.ListHostComponentsByService(service)
					componentMap := make(map[string]bool)
					for _, hostComponent := range hostComponents {
						componentMap[hostComponent.HostComponentName] = true
					}
					for component := range componentMap {
						componentFilter := Filter{Hosts: filter.Hosts, Components: []string{component}}
						hosts := a.GetFilteredHosts(componentFilter)
						componentDownloadFolder := createDownloadFolder(downloadFolder, component)
						a.CopyFolderFromRemote(component, componentLogDirMap[component], componentDownloadFolder, hosts, filter.Server)
					}
				}
			}
			if len(filter.Components) > 0 {
				for _, component := range filter.Components {
					componentFilter := Filter{Hosts: filter.Hosts, Components: []string{component}}
					hosts := a.GetFilteredHosts(componentFilter)
					componentDownloadFolder := createDownloadFolder(downloadFolder, component)
					a.CopyFolderFromRemote(component, componentLogDirMap[component], componentDownloadFolder, hosts, filter.Server)
				}
			}
		} else {
			hosts := a.GetFilteredHosts(filter)
			ambariAgentLogDir := "/var/log/ambari-agent"
			componentName := "ambari-agent"
			getLogDirCommand := "cat /etc/ambari-agent/conf/ambari-agent.ini | grep logdir"
			for host := range hosts {
				smallMap := make(map[string]bool)
				smallMap[host] = true
				responses := a.RunRemoteHostCommand(getLogDirCommand, smallMap, filter.Server)
				for _, response := range responses {
					splittedResponses := strings.Split(response.StdOut, "\n")
					propertyMap := ConvertStingsToMap(splittedResponses)
					ambariAgentLogDirValue := propertyMap["logdir"]
					ambariAgentLogDir = strings.TrimSpace(ambariAgentLogDirValue)
				}
				break
			}
			componentDownloadFolder := createDownloadFolder(downloadFolder, componentName)
			a.CopyFolderFromRemote(componentName, ambariAgentLogDir, componentDownloadFolder, hosts, filter.Server)
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
				findLogDirConfigsWithFilters(filter, components, blueprint, componentLogDirMap)
			}
		}
	}
	return componentLogDirMap
}

func findLogDirConfigsWithFilters(filter Filter, components map[string]map[string]string, blueprint map[string]interface{}, componentLogDirMap map[string]string) {
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

func createDownloadFolder(dest string, component string) string {
	newPath := path.Join(dest, component)
	os.MkdirAll(newPath, os.ModePerm)
	return newPath
}

func createDownloadRootFolder(dest string, registryName string) string {
	t := time.Now()
	timestamp := t.Format("20060102150405")
	newPath := path.Join(dest, "download-"+registryName+"-"+timestamp)
	os.MkdirAll(newPath, os.ModePerm)
	return newPath
}
