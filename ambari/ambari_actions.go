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

// ShowDetails get the registered Ambari server details
func (a AmbariRegistry) ShowDetails() {
	details := fmt.Sprintf("%s - %s://%s:%v - %s - %s / ********", a.name, a.protocol,
		a.hostname, a.port, a.cluster, a.username)
	fmt.Println(details)
}
