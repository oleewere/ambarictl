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
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ListAgents get all the registered hosts
func (a AmbariRegistry) ListAgents() []Host {
	request := a.CreateGetRequest("hosts?fields=Hosts/public_host_name,Hosts/ip,Hosts/host_state,Hosts/os_type,Hosts/os_arch,Hosts/last_agent_env", false)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Hosts
}

// ListServices get all installed services
func (a AmbariRegistry) ListServices() []Service {
	request := a.CreateGetRequest("services?fields=ServiceInfo/state,ServiceInfo/service_name", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Services
}

//ListComponents get all installed components
func (a AmbariRegistry) ListComponents() []Component {
	request := a.CreateGetRequest("components?fields=ServiceComponentInfo/component_name,ServiceComponentInfo/service_name,ServiceComponentInfo/state", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Components
}

//ListHostComponents get all installed host components by component type (or hosts)
func (a AmbariRegistry) ListHostComponents(param string, useHost bool) []HostComponent {
	var request *http.Request
	if useHost {
		request = a.CreateGetRequest("host_components?fields=HostRoles/component_name,HostRoles/state,HostRoles/host_name&HostRoles/host_name="+param, true)
	} else {
		request = a.CreateGetRequest("host_components?fields=HostRoles/component_name,HostRoles/state,HostRoles/host_name&HostRoles/component_name="+param, true)
	}
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().HostComponents
}

//ListHostComponentsByService get all installed host components by service name
func (a AmbariRegistry) ListHostComponentsByService(service string) []HostComponent {
	request := a.CreateGetRequest("host_components?fields=HostRoles/component_name,HostRoles/state,HostRoles/host_name&component/ServiceComponentInfo/service_name="+service, true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().HostComponents
}

// ListServiceConfigVersions gather service configuration details
func (a AmbariRegistry) ListServiceConfigVersions() []ServiceConfig {
	request := a.CreateGetRequest("configurations/service_config_versions?fields=service_name&is_current=true", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().ServiceConfigs
}

// GetClusterInfo obtain cluster detauls for ambari managed cluster
func (a AmbariRegistry) GetClusterInfo() Cluster {
	request := a.CreateGetRequest("?fields=Clusters/cluster_name,Clusters/version,Clusters/total_hosts,Clusters/security_type", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Cluster
}

// ExportBlueprint generate re-usable JSON from the cluster
func (a AmbariRegistry) ExportBlueprint() []byte {
	request := a.CreateGetRequest("?format=blueprint", true)
	return ProcessRequest(request)
}

// ExportBlueprintAsMap generate re-usable JSON map from the cluster
func (a AmbariRegistry) ExportBlueprintAsMap() map[string]interface{} {
	request := a.CreateGetRequest("?format=blueprint", true)
	return ProcessAsMap(request)
}

// GetStackDefaultConfigs obtain default configs for specific (versioned) stack
func (a AmbariRegistry) GetStackDefaultConfigs(stack string, version string) map[string]StackConfig {
	uriSuffix := fmt.Sprintf("stacks/%v/versions/%v/services?fields=configurations/*", stack, version)
	request := a.CreateGetRequest(uriSuffix, false)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().StackConfigs
}

// SetConfig sets a config value for a specific config key of a config type
func (a AmbariRegistry) SetConfig(configType string, configKey string, configValue string) {
	filter := Filter{}
	filter.Server = true
	filteredHosts := a.GetFilteredHosts(filter)
	versionNote := fmt.Sprintf("AMBARICTL - Update config key: %s", configKey)
	command := fmt.Sprintf("/var/lib/ambari-server/resources/scripts/configs.py --action set -c %s -k %s -v %s "+
		"-u %s -p %s --host=%s --cluster=%s --protocol=%s -b '%s'", configType, configKey, configValue, a.Username, a.Password,
		a.Hostname, a.Cluster, a.Protocol, versionNote)
	a.RunRemoteHostCommand(command, filteredHosts, filter.Server)
}

// RunAmbariServiceCommand start / stop / restart Ambari services or components
func (a AmbariRegistry) RunAmbariServiceCommand(command string, filter Filter, useServiceFilter bool, useComponentFilter bool) {
	command = strings.ToUpper(command)
	if command == "START" {
		a.startAmbariServiceOrComponent(useComponentFilter, filter, useServiceFilter)
	} else if command == "STOP" {
		a.stopAmbariServiceOrComponent(useComponentFilter, filter, useServiceFilter)
	} else if command == "RESTART" {
		a.restartAmbariServiceOrComponent(useComponentFilter, filter, useServiceFilter)
	} else {
		fmt.Println("Only START/STOP/RESTART operations are supported.")
		os.Exit(1)
	}
}

// StartService starting an ambari service
func (a AmbariRegistry) StartService(service string) []byte {
	request := a.serviceOperation(service, "STARTED", fmt.Sprintf("Start service (%s) by ambarictl", service))
	return ProcessRequest(request)
}

// StopService stopping an ambari service
func (a AmbariRegistry) StopService(service string) []byte {
	request := a.serviceOperation(service, "INSTALLED", fmt.Sprintf("Stop service (%s) by ambarictl", service))
	return ProcessRequest(request)
}

// RestartService restarting an ambari service
func (a AmbariRegistry) RestartService(service string) {
	a.StopService(service)
	a.StartService(service)
}

// StartComponent start an ambari component of a service
func (a AmbariRegistry) StartComponent(component string) []byte {
	request := a.componentOperation(component, "START", fmt.Sprintf("Start component (%s) by ambarictl", component))
	return ProcessRequest(request)
}

// StopComponent stop an ambari component of a service
func (a AmbariRegistry) StopComponent(component string) []byte {
	request := a.componentOperation(component, "STOP", fmt.Sprintf("Stop component (%s) by ambarictl", component))
	return ProcessRequest(request)
}

// RestartComponent restarts an ambari component of a service
func (a AmbariRegistry) RestartComponent(component string) []byte {
	request := a.componentOperation(component, "RESTART", fmt.Sprintf("Restart component (%s) by ambarictl", component))
	return ProcessRequest(request)
}

func getServiceNameForComponent(searchComponent string, components []Component) string {
	result := ""
	for _, component := range components {
		if component.ComponentName == searchComponent {
			return component.ServiceName
		}
	}
	return result
}

func (a AmbariRegistry) serviceOperation(service string, state string, context string) *http.Request {
	uriSuffix := fmt.Sprintf("services/%s", service)
	var bodyBytes bytes.Buffer
	jsonStr := fmt.Sprintf(`{"RequestInfo": {"context" : "%s"}, "Body": {"ServiceInfo": {"state": "%s"}}}`, context, state)
	bodyBytes.WriteString(jsonStr)
	return a.CreatePutRequest(bodyBytes, uriSuffix, true)
}

func (a AmbariRegistry) componentOperation(component string, operation string, context string) *http.Request {
	components := a.ListComponents()
	service := getServiceNameForComponent(component, components)
	hostComponents := a.ListHostComponents(component, false)
	hosts := ""
	for _, hostComponent := range hostComponents {
		hosts += hostComponent.HostComponntHost + ","
	}
	hosts = strings.TrimSuffix(hosts, ",")
	uriSuffix := "requests"
	var bodyBytes bytes.Buffer
	jsonStr := fmt.Sprintf(`{
  "RequestInfo": {
    "command": "%s",
    "context": "%s",
    "operation_level": {
      "level": "HOST_COMPONENT",
      "cluster_name": "%s"
    }
  },
  "Requests/resource_filters": [
    {
      "service_name": "%s",
      "component_name": "%s",
      "hosts": "%s"
    }
  ]
}`, operation, context, a.Cluster, service, component, hosts)
	bodyBytes.WriteString(jsonStr)
	return a.CreatePostRequest(bodyBytes, uriSuffix, true)
}

func (a AmbariRegistry) restartAmbariServiceOrComponent(useComponentFilter bool, filter Filter, useServiceFilter bool) {
	if useComponentFilter {
		for _, component := range filter.Components {
			a.RestartComponent(component)
		}
	} else if useServiceFilter {
		for _, component := range filter.Components {
			a.RestartService(component)
		}
	}
}

func (a AmbariRegistry) stopAmbariServiceOrComponent(useComponentFilter bool, filter Filter, useServiceFilter bool) {
	if useComponentFilter {
		for _, component := range filter.Components {
			a.StopComponent(component)
		}
	} else if useServiceFilter {
		for _, component := range filter.Services {
			a.StartService(component)
		}
	}
}

func (a AmbariRegistry) startAmbariServiceOrComponent(useComponentFilter bool, filter Filter, useServiceFilter bool) {
	if useComponentFilter {
		for _, component := range filter.Components {
			a.StartComponent(component)
		}
	} else if useServiceFilter {
		for _, component := range filter.Services {
			a.StartService(component)
		}
	}
}
