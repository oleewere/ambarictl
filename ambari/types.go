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

// AmbariRegistry represents registered ambari server entry details
type AmbariRegistry struct {
	name     string
	hostname string
	port     int
	username string
	password string
	protocol string
	cluster  string
	active   int
}

// AmbariItems global items from Ambari rest API response
type AmbariItems struct {
	Href  string `json:"href"`
	Items []Item `json:"items"`
}

// Item dynamic map - cast contents to specific types
type Item map[string]interface{}

// Host agent host details
type Host struct {
	HostName       string `json:"host_name,omitempty"`
	IP             string `json:"ip,omitempty"`
	PublicHostname string `json:"public_host_name,omitempty"`
	HostState      string `json:"host_state,omitempty"`
}

// Service ambari managed service info
type Service struct {
	ServiceName  string `json:"service_name,omitempty"`
	ServiceState string `json:"state,omitempty"`
}

// Component ambari managed component details
type Component struct {
	ComponentName  string `json:"component_name,omitempty"`
	ComponentState string `json:"state,omitempty"`
}

// Response common type which wraps all of the possible response entry types
type Response struct {
	Hosts      []Host
	Services   []Service
	Components []Component
}
