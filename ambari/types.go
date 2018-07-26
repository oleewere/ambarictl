package ambari

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

type AmbariItems struct {
	Href  string `json:"href"`
	Items []Item `json:"items"`
}

type Item map[string]interface{}

type Host struct {
	HostName       string `json:"host_name,omitempty"`
	IP             string `json:"ip,omitempty"`
	PublicHostname string `json:"public_host_name,omitempty"`
	HostState      string `json:"host_state,omitempty"`
}
