package ambari

import (
	"fmt"
)

// List agent hosts
func (a AmbariRegistry) ListAgents() []Host {
	request := a.CreateGetRequest("hosts?fields=Hosts/public_host_name,Hosts/ip,Hosts/host_state", false)
	ambariItems := ProcessAmbariItems(request)
	hosts := []Host{}
	for _, item := range ambariItems.Items {
		hostI := item["Hosts"].(map[string]interface{})
		host := Host{
			HostName:       hostI["host_name"].(string),
			IP:             hostI["ip"].(string),
			PublicHostname: hostI["public_host_name"].(string),
			HostState:      hostI["host_state"].(string),
		}
		hosts = append(hosts, host)
	}
	return hosts
}

// Show Ambari registry details
func (a AmbariRegistry) ShowDetails() {
	details := fmt.Sprintf("%s - %s://%s:%v - %s - %s / ********", a.name, a.protocol,
		a.hostname, a.port, a.cluster, a.username)
	fmt.Println(details)
}
