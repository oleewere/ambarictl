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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// CreateGetRequest creates an Ambari GET request
func (a AmbariRegistry) CreateGetRequest(urlSuffix string, useCluster bool) *http.Request {
	uri := a.GetAmbariUri(urlSuffix, useCluster)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.SetBasicAuth(a.Username, a.Password)
	return request
}

// CreatePostRequest creates an Ambari POST request with body
func (a AmbariRegistry) CreatePostRequest(body bytes.Buffer, urlSuffix string, useCluster bool) *http.Request {
	uri := a.GetAmbariUri(urlSuffix, useCluster)
	request, err := http.NewRequest("POST", uri, &body)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.SetBasicAuth(a.Username, a.Password)
	return request
}

// GetAmbariUri creates the Ambari uri with /api/v1/ suffix (+ /api/v1/clusters/<cluster> suffix is useCluster is enabled)
func (a AmbariRegistry) GetAmbariUri(uriSuffix string, useCluster bool) string {
	if useCluster {
		return fmt.Sprintf("%s://%s:%v/api/v1/clusters/%s/%s", a.Protocol, a.Hostname, a.Port, a.Cluster, uriSuffix)
	}
	return fmt.Sprintf("%s://%s:%v/api/v1/%s", a.Protocol, a.Hostname, a.Port, uriSuffix)
}

// GetHttpClient create HTTP client instance for Ambari
func GetHttpClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}
	return httpClient
}

// ProcessAmbariItems get "items" from Ambari response
func ProcessAmbariItems(request *http.Request) AmbariItems {
	bodyBytes := ProcessRequest(request)
	var ambariItems AmbariItems
	err := json.Unmarshal(bodyBytes, &ambariItems)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return ambariItems
}

// ProcessRequest get a simple response from a REST call
func ProcessRequest(request *http.Request) []byte {
	client := GetHttpClient()
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return bodyBytes
}
