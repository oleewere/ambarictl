package ambari

import (
	"fmt"
)

// List agent hosts
func (a AmbariRegistry) ListAgents() []Host {
	request := a.CreateGetRequest("hosts?fields=Hosts/public_host_name,Hosts/ip,Hosts/host_state", false)
	ambariItems := ProcessAmbariItems(request)
	return ambariItems.ConvertResponse().Hosts
}

// Show Ambari registry details
func (a AmbariRegistry) ShowDetails() {
	details := fmt.Sprintf("%s - %s://%s:%v - %s - %s / ********", a.name, a.protocol,
		a.hostname, a.port, a.cluster, a.username)
	fmt.Println(details)
}
