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

import "strings"

// Filter represents filter on agent hosts (by component / service / hosts)
type Filter struct {
	Services   []string
	Components []string
	Hosts      []string
	Server     bool
}

// CreateFilter will make a Filter object from filter strings (component / service / hosts)
func CreateFilter(serviceFilter string, componentFilter string, hostFilter string, ambariServer bool) Filter {
	filter := Filter{}
	if len(serviceFilter) > 0 {
		services := strings.Split(serviceFilter, ",")
		filter.Services = services
	}
	if len(componentFilter) > 0 {
		components := strings.Split(componentFilter, ",")
		filter.Components = components
	}
	if len(hostFilter) > 0 {
		hosts := strings.Split(hostFilter, ",")
		filter.Hosts = hosts
	}
	filter.Server = ambariServer
	return filter
}

// GetFilteredHosts obtain specific hosts based on different filters
func (a AmbariRegistry) GetFilteredHosts(filter Filter) map[string]bool {
	finalHosts := make(map[string]bool)
	hosts := make(map[string]bool) // use boolean map as a set
	if len(filter.Services) > 0 {
		for _, service := range filter.Services {
			hostComponents := a.ListHostComponentsByService(service)
			for _, hostComponent := range hostComponents {
				hosts[hostComponent.HostComponntHost] = true
			}
		}
	}
	if len(filter.Components) > 0 {
		for _, component := range filter.Components {
			hostComponents := a.ListHostComponents(component, false)
			for _, hostComponent := range hostComponents {
				hosts[hostComponent.HostComponntHost] = true
			}
		}
	}
	if filter.Server {
		hosts[a.Hostname] = true
	}
	agents := a.ListAgents()
	for _, agent := range agents {
		if len(filter.Hosts) > 0 {
			filteredHosts := filter.Hosts
			containsHost := false
			for _, filteredHost := range filteredHosts {
				if filteredHost == agent.PublicHostname {
					containsHost = true
				}
			}
			if !containsHost {
				continue
			}
		}
		if len(hosts) > 0 {
			_, ok := hosts[agent.PublicHostname]
			if ok {
				finalHosts[agent.IP] = true
			}
			_, ok = hosts[agent.IP]
			if ok {
				finalHosts[agent.IP] = true
			}

		} else {
			finalHosts[agent.IP] = true
		}
	}
	return finalHosts
}
