package ambari

// Convert response items to specific types
func (a AmbariItems) ConvertResponse() Response {
	response := Response{}
	hosts := []Host{}
	for _, item := range a.Items {
		hosts = createHostsType(item, hosts)
	}
	if len(hosts) > 0 {
		response.Hosts = hosts
	}
	return response
}

func createHostsType(item Item, hosts []Host) []Host {
	if hosts_val, ok := item["Hosts"]; ok {
		host := Host{}
		hostI := hosts_val.(map[string]interface{})
		if hostname, ok := hostI["host_name"]; ok {
			host.HostName = hostname.(string)
		}
		if ip, ok := hostI["ip"]; ok {
			host.IP = ip.(string)
		}
		if public_host_name, ok := hostI["public_host_name"]; ok {
			host.PublicHostname = public_host_name.(string)
		}
		if host_state, ok := hostI["host_state"]; ok {
			host.HostState = host_state.(string)
		}

		hosts = append(hosts, host)
	}
	return hosts
}
