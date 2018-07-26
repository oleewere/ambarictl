package ambari

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
)

// List agent hosts
func (a AmbariRegistry) ListAgents() {
	client := GetHttpClient()
	request := a.CreateGetRequest("hosts?fields=Hosts/public_host_name,Hosts/ip,Hosts/host_state", false)

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	var ambariItems AmbariItems
	json_err := json.Unmarshal(bodyBytes, &ambariItems)
	if json_err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	hosts := []Host{}
	for _, item := range ambariItems.Items {
		hostI := item["Hosts"].(map[string]interface{})
		host := Host{
			HostName: hostI["host_name"].(string),
			IP: hostI["ip"].(string),
			PublicHostname: hostI["public_host_name"].(string),
			HostState: hostI["host_state"].(string),
		}
		hosts = append(hosts, host)
	}

	fmt.Println("Registered hosts:")
	fmt.Println("-----------------")
	for _, host := range hosts {
		hostEntry := fmt.Sprintf("%s (ip: %s) - state: %s", host.PublicHostname, host.IP, host.HostState)
		fmt.Println(hostEntry)
	}

}

// Show Ambari registry details
func (a AmbariRegistry) ShowDetails() {
	details := fmt.Sprintf("%s - %s://%s:%v - %s - %s / ********", a.name, a.protocol,
		a.hostname, a.port, a.cluster, a.username)
	fmt.Println(details)
}
