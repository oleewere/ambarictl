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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const (
	ambariServerJsonFileName       = "ambari_servers.json"
	connectionProfilesJsonFileName = "connection_profiles.json"
)

// CreateAmbariRegistryDb initialize ambarictl database
func CreateAmbariRegistryDb() {
	ambariServerJsonFile := getJsonDbFile(ambariServerJsonFileName)
	connectionProfileJsonFile := getJsonDbFile(connectionProfilesJsonFileName)
	if !exists(ambariServerJsonFile) {
		ambariServerRegistries := make([]AmbariRegistry, 0)
		ambariServerJson, _ := json.Marshal(ambariServerRegistries)
		err := ioutil.WriteFile(ambariServerJsonFile, ambariServerJson, 0644)
		checkErr(err)
	}
	if !exists(connectionProfileJsonFile) {
		connectionProfiles := make([]ConnectionProfile, 0)
		connectionProfilesJson, _ := json.Marshal(connectionProfiles)
		err := ioutil.WriteFile(connectionProfileJsonFile, connectionProfilesJson, 0644)
		checkErr(err)
	}
}

// DropAmbariRegistryRecords drop all ambari server entries from ambarictl database
func DropAmbariRegistryRecords() {
	ambariServerRegistries := make([]AmbariRegistry, 0)
	WriteAmbariServerEntries(ambariServerRegistries)
}

// DropConnectionProfileRecords drop all connection profile from ambarictl database
func DropConnectionProfileRecords() {
	connectionProfiles := make([]ConnectionProfile, 0)
	WriteConnectionProfileEntries(connectionProfiles)
}

// ListAmbariRegistryEntries get all ambari registries from ambarictl database
func ListAmbariRegistryEntries() []AmbariRegistry {
	ambariServerJsonFile := getJsonDbFile(ambariServerJsonFileName)
	file, err := ioutil.ReadFile(ambariServerJsonFile)
	checkErr(err)
	ambariRegistries := make([]AmbariRegistry, 0)
	json.Unmarshal(file, &ambariRegistries)
	return ambariRegistries
}

// ListConnectionProfileEntries get all ambari registries from ambarictl database
func ListConnectionProfileEntries() []ConnectionProfile {
	connectionProfileJsonFile := getJsonDbFile(connectionProfilesJsonFileName)
	file, err := ioutil.ReadFile(connectionProfileJsonFile)
	checkErr(err)
	connectionProfiles := make([]ConnectionProfile, 0)
	json.Unmarshal(file, &connectionProfiles)
	return connectionProfiles
}

// GetAmbariEntryId get ambari entry id if the id exists
func GetAmbariEntryId(id string) string {
	ambariEntries := ListAmbariRegistryEntries()
	ambariEntryId := ""
	if len(ambariEntries) > 0 {
		for _, ambariEntry := range ambariEntries {
			if ambariEntry.Name == id {
				ambariEntryId = ambariEntry.Name
			}
		}
	}
	return ambariEntryId
}

// GetConnectionProfileEntryId get connection profile entry id if the id exists
func GetConnectionProfileEntryId(id string) string {
	connectionProfiles := ListConnectionProfileEntries()
	connectionProfileId := ""
	if len(connectionProfiles) > 0 {
		for _, connectionProfileEntry := range connectionProfiles {
			if connectionProfileEntry.Name == id {
				connectionProfileId = connectionProfileEntry.Name
			}
		}
	}
	return connectionProfileId
}

// RegisterNewAmbariEntry create new ambari registry entry in ambarictl database
func RegisterNewAmbariEntry(id string, hostname string, port int, protocol string, username string, password string, cluster string) {
	checkId := GetAmbariEntryId(id)
	if len(checkId) > 0 {
		alreadyExistMsg := fmt.Sprintf("Registry with id '%s' is already defined as a registry entry", checkId)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}
	ambaiServerEntries := ListAmbariRegistryEntries()
	newAmbariServerEntry := AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol, Username: username, Password: password, Cluster: cluster, Active: true}
	ambaiServerEntries = append(ambaiServerEntries, newAmbariServerEntry)
	WriteAmbariServerEntries(ambaiServerEntries)
}

// RegisterNewConnectionProfile create new connection profile entry in ambarictl database
func RegisterNewConnectionProfile(id string, keyPath string, port int, username string, hostJump bool, proxyAddress string) {
	checkId := GetConnectionProfileEntryId(id)
	if len(checkId) > 0 {
		alreadyExistMsg := fmt.Sprintf("Connection profile with id '%s' is already defined as a profile entry", checkId)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}
	connectionProfiles := ListConnectionProfileEntries()
	newConnectionProfile := ConnectionProfile{Name: id, KeyPath: keyPath, Port: port, Username: username, HostJump: hostJump, ProxyAddress: proxyAddress}
	connectionProfiles = append(connectionProfiles, newConnectionProfile)
	WriteConnectionProfileEntries(connectionProfiles)
}

// DeRegisterAmbariEntry remove an ambari server enrty by id
func DeRegisterAmbariEntry(id string) {
	ambariServers := ListAmbariRegistryEntries()
	newAmbariServers := make([]AmbariRegistry, 0)
	if len(ambariServers) > 0 {
		for index := range ambariServers {
			if ambariServers[index].Name != id {
				newAmbariServers = append(newAmbariServers, ambariServers[index])
			}
		}
	}
	WriteAmbariServerEntries(ambariServers)
}

// DeRegisterConnectionProfile remove a connection profile by id
func DeRegisterConnectionProfile(id string) {
	connectionProfiles := ListConnectionProfileEntries()
	newConnectionProfiles := make([]ConnectionProfile, 0)
	if len(connectionProfiles) > 0 {
		for index := range connectionProfiles {
			if connectionProfiles[index].Name != id {
				newConnectionProfiles = append(newConnectionProfiles, connectionProfiles[index])
			}
		}
	}
	WriteConnectionProfileEntries(newConnectionProfiles)
}

// GetActiveAmbari get the active ambari registry from ambarictl database (should be only one)
func GetActiveAmbari() AmbariRegistry {
	ambariServers := ListAmbariRegistryEntries()
	var result AmbariRegistry
	if len(ambariServers) > 0 {
		for _, ambariServerEntry := range ambariServers {
			if ambariServerEntry.Active {
				result = ambariServerEntry
			}
		}
	}
	return result
}

// GetAmbariById get the ambari registry from ambarictl database by id
func GetAmbariById(searchId string) AmbariRegistry {
	ambariServers := ListAmbariRegistryEntries()
	var result AmbariRegistry
	if len(ambariServers) > 0 {
		for _, ambariServerEntry := range ambariServers {
			if ambariServerEntry.Name == searchId {
				result = ambariServerEntry
			}
		}
	}
	return result
}

// GetConnectionProfileById get the connection profile from ambarictl database by id
func GetConnectionProfileById(searchId string) ConnectionProfile {
	connectionProfiles := ListConnectionProfileEntries()
	var result ConnectionProfile
	if len(connectionProfiles) > 0 {
		for _, connectionProfileEntry := range connectionProfiles {
			if connectionProfileEntry.Name == searchId {
				result = connectionProfileEntry
			}
		}
	}
	return result
}

// SetProfileIdForAmbariEntry attach a connection profile to a specific ambari server entry
func SetProfileIdForAmbariEntry(ambariEntryId string, profileId string) {
	ambariServers := ListAmbariRegistryEntries()
	if len(ambariServers) > 0 {
		for index := range ambariServers {
			if ambariServers[index].Name == ambariEntryId {
				ambariServers[index].ConnectionProfile = profileId
			}
		}
	}
	WriteAmbariServerEntries(ambariServers)
}

// ActiveAmbariRegistry turn on active status on selected ambari registry
func ActiveAmbariRegistry(id string) {
	checkId := GetAmbariEntryId(id)
	if len(checkId) == 0 {
		alreadyExistMsg := fmt.Sprintf("Not found Ambari server registry  with id '%s'.", checkId)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}
	ambariServers := ListAmbariRegistryEntries()
	if len(ambariServers) > 0 {
		for index := range ambariServers {
			if ambariServers[index].Name == id {
				ambariServers[index].Active = true
			} else {
				ambariServers[index].Active = false
			}
		}
	}
	WriteAmbariServerEntries(ambariServers)
}

// DeactiveAllAmbariRegistry turn off active status on all ambari registries
func DeactiveAllAmbariRegistry() {
	ambariServers := ListAmbariRegistryEntries()
	if len(ambariServers) > 0 {
		for index := range ambariServers {
			ambariServers[index].Active = false
		}
	}
	WriteAmbariServerEntries(ambariServers)
}

// WriteAmbariServerEntries write ambari server entries to the ambari server registry json file
func WriteAmbariServerEntries(ambariServers []AmbariRegistry) {
	ambariServerJson, _ := json.Marshal(ambariServers)
	ambariServerJsonFile := getJsonDbFile(ambariServerJsonFileName)
	err := ioutil.WriteFile(ambariServerJsonFile, FormatJson(ambariServerJson).Bytes(), 0600)
	checkErr(err)
}

// WriteConnectionProfileEntries write connection profile entries to the connection profile registry json file
func WriteConnectionProfileEntries(connectionProfiles []ConnectionProfile) {
	connectionProfilesJson, _ := json.Marshal(connectionProfiles)
	connectionProfilesJsonFile := getJsonDbFile(connectionProfilesJsonFileName)
	err := ioutil.WriteFile(connectionProfilesJsonFile, FormatJson(connectionProfilesJson).Bytes(), 0600)
	checkErr(err)
}

// FormatJson format json file
func FormatJson(b []byte) *bytes.Buffer {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &out
}

func getJsonDbFile(file string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	home := usr.HomeDir
	ambariManagerFolder := path.Join(home, ".ambarictl")
	if _, err := os.Stat(ambariManagerFolder); os.IsNotExist(err) {
		os.Mkdir(ambariManagerFolder, os.ModePerm)
	}
	return path.Join(ambariManagerFolder, file)
}

// Exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
