package ambari

import (
	"net/http"
	"bytes"
	"fmt"
	"time"
	"crypto/tls"
)

// Create Ambari GET request
func (a AmbariRegistry) CreateGetRequest(urlSuffix string, useCluster bool) *http.Request {
	uri := a.GetAmbariUri(urlSuffix, useCluster)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.SetBasicAuth(a.username, a.password)
	return request
}

// Create Ambari POST request with body
func (a AmbariRegistry) CreatePostRequest(body bytes.Buffer, urlSuffix string, useCluster bool) *http.Request {
	uri := a.GetAmbariUri(urlSuffix, useCluster)
	request, err := http.NewRequest("POST", uri, &body)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.SetBasicAuth(a.username, a.password)
	return request
}

// Create Ambari uri with /api/v1/ suffix (+ /api/v1/clusters/<cluster> suffix is useCluster is enabled)
func (a AmbariRegistry) GetAmbariUri(uriSuffix string, useCluster bool) string {
	if useCluster {
		uri := fmt.Sprintf("%s://%s:%v/api/v1/clusters/%s/%s", a.protocol, a.hostname, a.port, a.cluster, uriSuffix)
		return uri
	} else {
		uri := fmt.Sprintf("%s://%s:%v/api/v1/%s", a.protocol, a.hostname, a.port, uriSuffix)
		return uri
	}
}

// Create Http client for Ambari
func GetHttpClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return httpClient
}
