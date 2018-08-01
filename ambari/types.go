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
	Name     string
	Hostname string
	Port     int
	Username string
	Password string
	Protocol string
	Cluster  string
	Active   int
}

// AmbariItems global items from Ambari rest API response
type AmbariItems struct {
	Href  string `json:"href"`
	Items []Item `json:"items"`
	Cluster Cluster `json:"Clusters,omitempty"`
}

// Item dynamic map - cast contents to specific types
type Item map[string]interface{}

// Host agent host details
type Host struct {
	HostName       string `json:"host_name,omitempty"`
	IP             string `json:"ip,omitempty"`
	PublicHostname string `json:"public_host_name,omitempty"`
	OSType         string `json:"os_type,omitempty"`
	OSArch         string `json:"os_arch,omitempty"`
	UnlimitedJCE   bool   `json:"unlimited_jce,omitempty"`
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

// HostComponent ambari managed host component details
type HostComponent struct {
	HostComponentName  string `json:"host_component_name,omitempty"`
	HostComponentState string `json:"state,omitempty"`
	HostComponntHost   string `json:"host_name,omitempty"`
}

// ServiceConfig represents service specific configurations
type ServiceConfig struct {
	ServiceConfigType    string     `json:"type,omitempty"`
	ServiceConfigTag     string     `json:"tag,omitempty"`
	ServiceConfigVersion float64    `json:"tag,omitempty"`
	Properties           Properties `json:"properties,omitempty"`
}

// Cluster holds installed ambari cluster details
type Cluster struct {
	ClusterName         string  `json:"cluster_name,omitempty"`
	ClusterVersion      string  `json:"version,omitempty"`
	ClusterTotalHosts   float64 `json:"total_hosts,omitempty"`
	ClusterSecurityType string  `json:"security_type,omitempty"`
}

// Properties represents configuration properties (key/value pairs)
type Properties map[string]interface{}

// Response common type which wraps all of the possible response entry types
type Response struct {
	Cluster        Cluster
	Hosts          []Host
	Services       []Service
	Components     []Component
	HostComponents []HostComponent
	ServiceConfigs []ServiceConfig
}
