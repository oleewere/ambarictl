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

// ConvertResponse converts the response items to specific types
func (a AmbariItems) ConvertResponse() Response {
	response := Response{}
	hosts := []Host{}
	services := []Service{}
	components := []Component{}
	hostComponents := []HostComponent{}
	for _, item := range a.Items {
		hosts = createHostsType(item, hosts)
		services = createServicesType(item, services)
		components = createComponentsType(item, components)
		hostComponents = createHostComponentsType(item, hostComponents)
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
	return response
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
