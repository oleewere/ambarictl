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
	"strings"
)

// ConvertResponse converts the response items to specific types
func (a AmbariItems) ConvertResponse() Response {
	response := Response{}
	hosts := []Host{}
	services := []Service{}
	components := []Component{}
	hostComponents := []HostComponent{}
	serviceConfigs := []ServiceConfig{}
	clusterInfo := Cluster{}
	clusterInfo = a.Cluster
	stackConfigs := make(map[string]StackConfig)
	for _, item := range a.Items {
		hosts = createHostsType(item, hosts)
		services = createServicesType(item, services)
		components = createComponentsType(item, components)
		hostComponents = createHostComponentsType(item, hostComponents)
		serviceConfigs = createServiceConfigsType(item, serviceConfigs)
		stackConfigs = createStackConfigsType(item, stackConfigs)
	}
	if len(hosts) > 0 {
		response.Hosts = hosts
	}
	if len(services) > 0 {
		response.Services = services
	}
	if len(components) > 0 {
		response.Components = components
	}
	if len(hostComponents) > 0 {
		response.HostComponents = hostComponents
	}
	if len(serviceConfigs) > 0 {
		response.ServiceConfigs = serviceConfigs
	}
	if len(clusterInfo.ClusterName) > 0 {
		response.Cluster = clusterInfo
	}
	if len(stackConfigs) > 0 {
		response.StackConfigs = stackConfigs
	}
	return response
}

func createServiceConfigsType(item Item, configs []ServiceConfig) []ServiceConfig {
	if configsVal, ok := item["configurations"]; ok {
		serviceConfI := configsVal.([]interface{})
		for _, configVal := range serviceConfI {
			serviceConfig := ServiceConfig{}
			confI := configVal.(map[string]interface{})
			if tag, ok := confI["tag"]; ok {
				serviceConfig.ServiceConfigTag = tag.(string)
			}
			if serviceConfigType, ok := confI["type"]; ok {
				serviceConfig.ServiceConfigType = serviceConfigType.(string)
			}
			if version, ok := confI["version"]; ok {
				serviceConfig.ServiceConfigVersion = version.(float64)
			}
			configs = append(configs, serviceConfig)
		}
	}
	return configs
}

func createHostsType(item Item, hosts []Host) []Host {
	if hostsVal, ok := item["Hosts"]; ok {
		host := Host{}
		hostI := hostsVal.(map[string]interface{})
		if hostName, ok := hostI["host_name"]; ok {
			host.HostName = hostName.(string)
		}
		if ip, ok := hostI["ip"]; ok {
			host.IP = ip.(string)
		}
		if publicHostName, ok := hostI["public_host_name"]; ok {
			host.PublicHostname = publicHostName.(string)
		}
		if hostState, ok := hostI["host_state"]; ok {
			host.HostState = hostState.(string)
		}
		if osType, ok := hostI["os_type"]; ok {
			host.OSType = osType.(string)
		}
		if osArch, ok := hostI["os_arch"]; ok {
			host.OSArch = osArch.(string)
		}
		if lastAgentEnvVal, ok := hostI["last_agent_env"]; ok {
			lastAgentEnv := lastAgentEnvVal.(map[string]interface{})
			if jceVal, ok := lastAgentEnv["hasUnlimitedJcePolicy"]; ok {
				host.UnlimitedJCE = jceVal.(bool)
			}
		}

		hosts = append(hosts, host)
	}
	return hosts
}

func createComponentsType(item Item, components []Component) []Component {
	if componentVal, ok := item["ServiceComponentInfo"]; ok {
		component := Component{}
		componentI := componentVal.(map[string]interface{})
		if componentName, ok := componentI["component_name"]; ok {
			component.ComponentName = componentName.(string)
		}
		if serviceName, ok := componentI["service_name"]; ok {
			component.ServiceName = serviceName.(string)
		}
		if state, ok := componentI["state"]; ok {
			component.ComponentState = state.(string)
		}

		components = append(components, component)
	}
	return components
}

func createHostComponentsType(item Item, hostComponents []HostComponent) []HostComponent {
	if hostComponentVal, ok := item["HostRoles"]; ok {
		hostComponent := HostComponent{}
		hostComponentI := hostComponentVal.(map[string]interface{})
		if hostComponentName, ok := hostComponentI["component_name"]; ok {
			hostComponent.HostComponentName = hostComponentName.(string)
		}
		if hostName, ok := hostComponentI["host_name"]; ok {
			hostComponent.HostComponntHost = hostName.(string)
		}
		if state, ok := hostComponentI["state"]; ok {
			hostComponent.HostComponentState = state.(string)
		}
		hostComponents = append(hostComponents, hostComponent)
	}
	return hostComponents
}

func createServicesType(item Item, services []Service) []Service {
	if servicesVal, ok := item["ServiceInfo"]; ok {
		service := Service{}
		serviceI := servicesVal.(map[string]interface{})
		if serviceName, ok := serviceI["service_name"]; ok {
			service.ServiceName = serviceName.(string)
		}
		if serviceState, ok := serviceI["state"]; ok {
			service.ServiceState = serviceState.(string)
		}
		services = append(services, service)
	}
	return services
}

func createStackConfigsType(item Item, stackConfigMap map[string]StackConfig) map[string]StackConfig {
	if configsVal, ok := item["configurations"]; ok {
		stackConfI := configsVal.([]interface{})
		for _, configVal := range stackConfI {
			confI := configVal.(map[string]interface{})
			if stackConfigPropertyMapVal, ok := confI["StackConfigurations"]; ok {
				stackConfigPropsMap := stackConfigPropertyMapVal.(map[string]interface{})
				stackConfigProp := createStackProperty(stackConfigPropsMap)
				stackConfigEntry, ok := stackConfigMap[stackConfigProp.Type]
				if ok {
					stackConfigEntry.Properties = append(stackConfigEntry.Properties, stackConfigProp)
					stackConfigMap[stackConfigProp.Type] = stackConfigEntry
				} else {
					stackConfig := StackConfig{}
					stackConfig.ServiceConfigType = stackConfigProp.Type
					stackConfig.Properties = append(stackConfig.Properties, stackConfigProp)
					stackConfigMap[stackConfigProp.Type] = stackConfig
				}
			}
		}
	}
	return stackConfigMap
}

func createStackProperty(stackConfigPropsMap map[string]interface{}) StackProperty {
	stackProperty := StackProperty{}
	if propertyName, ok := stackConfigPropsMap["property_name"]; ok {
		stackProperty.Name = propertyName.(string)
	}
	if propertyValue, ok := stackConfigPropsMap["property_value"]; ok {
		if propertyValue == nil {
			stackProperty.Value = ""
		} else {
			stackProperty.Value = propertyValue.(string)
		}
	}
	if propertyTypeVal, ok := stackConfigPropsMap["property_type"]; ok {
		propertyTypeSlice := propertyTypeVal.([]interface{})
		if len(propertyTypeSlice) < 0 {
			propertyType := propertyTypeSlice[0]
			stackProperty.PropertyType = propertyType.(string)
		}
	}
	if typeVal, ok := stackConfigPropsMap["type"]; ok {
		typeValue := strings.TrimSuffix(typeVal.(string), ".xml")
		stackProperty.Type = typeValue
	}
	return stackProperty
}
