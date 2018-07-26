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
	for _, item := range a.Items {
		hosts = createHostsType(item, hosts)
		services = createServicesType(item, services)
	}
	if len(hosts) > 0 {
		response.Hosts = hosts
	}
	if len(services) > 0 {
		response.Services = services
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
