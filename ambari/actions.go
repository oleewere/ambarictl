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

import "net/http"

// ListAgents get all the registered hosts
func (a AmbariRegistry) ListAgents() []Host {
	request := a.CreateGetRequest("hosts?fields=Hosts/public_host_name,Hosts/ip,Hosts/host_state", false)
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
	request := a.CreateGetRequest("components?fields=ServiceComponentInfo/component_name,ServiceComponentInfo/state", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Components
}

//ListHostComponents get all installed host components by component type
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

// ListServiceConfigVersions gather service configuration details
func (a AmbariRegistry) ListServiceConfigVersions() []ServiceConfig {
	request := a.CreateGetRequest("configurations/service_config_versions?fields=service_name&is_current=true", true)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().ServiceConfigs
}
