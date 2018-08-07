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
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"os"
	"os/user"
	"path"
	"strconv"
)

// CreateAmbariRegistryDb initialize ambarictl database
func CreateAmbariRegistryDb() {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS connection_profile (id VARCHAR PRIMARY KEY, port INTEGER, key_path VARCHAR, username VARCHAR, host_jump INTEGER, proxy_address VARCHAR)")
	checkErr(err)
	_, connProfErr := statement.Exec()
	checkErr(connProfErr)
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS ambari_registry (id VARCHAR PRIMARY KEY, hostname VARCHAR, port INTEGER, protocol VARCHAR, username VARCHAR, password VARCHAR, cluster TEXT, active INTEGER, connection_profile VARCHAR DEFAULT '' REFERENCES connection_profile(id) ON DELETE SET DEFAULT)")
	checkErr(err)
	_, regErr := statement.Exec()
	checkErr(regErr)
}

// DropAmbariRegistryRecords drop all ambari server entries from ambarictl database
func DropAmbariRegistryRecords() {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("DELETE from ambari_registry")
	checkErr(err)
	statement.Exec()
}

// DropConnectionProfileRecords drop all connection profile from ambarictl database
func DropConnectionProfileRecords() {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("DELETE from connection_profile")
	checkErr(err)
	statement.Exec()
	statement, err = db.Prepare("UPDATE ambari_registry SET connection_profile=''")
	checkErr(err)
	statement.Exec()
}

// ListAmbariRegistryEntries get all ambari registries from ambarictl database
func ListAmbariRegistryEntries() []AmbariRegistry {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,password,cluster,active,connection_profile FROM ambari_registry")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var password string
	var cluster string
	var active int
	var connectionProfile string
	var ambariRegistries []AmbariRegistry
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &password, &cluster, &active, &connectionProfile)
		ambariRegistry := AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol,
			Username: username, Password: password, Cluster: cluster, Active: active, ConnectionProfile: connectionProfile}
		ambariRegistries = append(ambariRegistries, ambariRegistry)
	}
	rows.Close()
	return ambariRegistries
}

// ListConnectionProfileEntries get all ambari registries from ambarictl database
func ListConnectionProfileEntries() []ConnectionProfile {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,port,key_path,username,host_jump,proxy_address FROM connection_profile")
	checkErr(err)
	var id string
	var keyPath string
	var port int
	var username string
	var hostJump int
	var proxyAddress string
	var connectionProfiles []ConnectionProfile
	for rows.Next() {
		rows.Scan(&id, &port, &keyPath, &username, &hostJump, &proxyAddress)
		connectionProfile := ConnectionProfile{Name: id, KeyPath: keyPath, Port: port, Username: username, HostJump: hostJump, ProxyAddress: proxyAddress}
		connectionProfiles = append(connectionProfiles, connectionProfile)
	}
	rows.Close()
	return connectionProfiles
}

// GetAmbariEntryId get ambari entry id if the id exists
func GetAmbariEntryId(id string) string {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id FROM ambari_registry WHERE id = '" + id + "'")
	checkErr(err)
	var ambariEntryId string
	for rows.Next() {
		rows.Scan(&ambariEntryId)
	}
	rows.Close()
	return ambariEntryId
}

// GetConnectionProfileEntryId get connection profile entry id if the id exists
func GetConnectionProfileEntryId(id string) string {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id FROM connection_profile WHERE id = '" + id + "'")
	checkErr(err)
	var connProfileId string
	for rows.Next() {
		rows.Scan(&connProfileId)
	}
	rows.Close()
	return connProfileId
}

// RegisterNewAmbariEntry create new ambari registry entry in ambarictl database
func RegisterNewAmbariEntry(id string, hostname string, port int, protocol string, username string, password string, cluster string) {
	checkId := GetAmbariEntryId(id)
	if len(checkId) > 0 {
		alreadyExistMsg := fmt.Sprintf("Registry with id '%s' is already defined as a registry entry", checkId)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("INSERT INTO ambari_registry (id, hostname, port, protocol, username, password, cluster, active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	_, insertErr := statement.Exec(id, hostname, strconv.Itoa(port), protocol, username, password, cluster, strconv.Itoa(1))
	checkErr(insertErr)
}

// RegisterNewConnectionProfile create new connection profile entry in ambarictl database
func RegisterNewConnectionProfile(id string, keyPath string, port int, username string, hostJump int, proxyAddress string) {
	checkId := GetConnectionProfileEntryId(id)
	if len(checkId) > 0 {
		alreadyExistMsg := fmt.Sprintf("Connection profile with id '%s' is already defined as a profile entry", checkId)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("INSERT INTO connection_profile (id, key_path, port, username, host_jump, proxy_address) VALUES (?, ?, ?, ?, ?, ?)")
	_, insertErr := statement.Exec(id, keyPath, strconv.Itoa(port), username, strconv.Itoa(hostJump), proxyAddress)
	checkErr(insertErr)
}

// DeRegisterAmbariEntry remove an ambari server enrty by id
func DeRegisterAmbariEntry(id string) {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("DELETE FROM ambari_registry WHERE id = ?")
	_, deleteErr := statement.Exec(id)
	checkErr(deleteErr)
}

// DeRegisterConnectionProfile remove a connection profile by id
func DeRegisterConnectionProfile(id string) {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("DELETE FROM connection_profile WHERE id = ?")
	_, deleteErr := statement.Exec(id)
	checkErr(deleteErr)
	statement, err = db.Prepare("UPDATE ambari_registry SET connection_profile='' WHERE connection_profile = ?")
	checkErr(err)
	statement.Exec(id)
}

// GetActiveAmbari get the active ambari registry from ambarictl database (should be only one)
func GetActiveAmbari() AmbariRegistry {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,password,cluster,connection_profile FROM ambari_registry WHERE active = '1'")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var password string
	var cluster string
	var connectionProfile string
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &password, &cluster, &connectionProfile)
	}
	rows.Close()

	return AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol, Username: username, Password: password, Cluster: cluster, Active: 1, ConnectionProfile: connectionProfile}
}

// GetAmbariById get the ambari registry from ambarictl database by id
func GetAmbariById(searchId string) AmbariRegistry {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,password,cluster,connection_profile FROM ambari_registry WHERE id = '" + searchId + "'")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var password string
	var cluster string
	var connectionProfile string
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &password, &cluster, &connectionProfile)
	}
	rows.Close()

	return AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol, Username: username, Password: password, Cluster: cluster, Active: 1, ConnectionProfile: connectionProfile}
}

// GetConnectionProfileById get the connection profile from ambarictl database by id
func GetConnectionProfileById(searchId string) ConnectionProfile {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,key_path,port,username,host_jump,proxy_address FROM connection_profile WHERE id = '" + searchId + "'")
	checkErr(err)
	var id string
	var keyPath string
	var port int
	var username string
	var hostJump int
	var proxyAddress string
	for rows.Next() {
		rows.Scan(&id, &keyPath, &port, &username, &hostJump, &proxyAddress)
	}
	rows.Close()

	return ConnectionProfile{Name: id, KeyPath: keyPath, Port: port, Username: username, HostJump: hostJump, ProxyAddress: proxyAddress}
}

// SetProfileIdForAmbariEntry attach a connection profile to a specific ambari server entry
func SetProfileIdForAmbariEntry(ambariEntryId string, profileId string) {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("UPDATE ambari_registry SET connection_profile=? WHERE id = ?")
	_, updateErr := statement.Exec(profileId, ambariEntryId)
	checkErr(updateErr)
}

// ActiveAmbariRegistry turn on active status on selected ambari registry
func ActiveAmbariRegistry(id string) {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("UPDATE ambari_registry SET active='1' WHERE id = ?")
	_, updateErr := statement.Exec(id)
	checkErr(updateErr)
}

// DeactiveAllAmbariRegistry turn off active status on all ambari registries
func DeactiveAllAmbariRegistry() {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	statement, _ := db.Prepare("UPDATE ambari_registry SET active='0' WHERE active = '1'")
	_, updateErr := statement.Exec()
	checkErr(updateErr)
}

func getDb() (*sql.DB, error) {
	drivers := sql.Drivers()
	driverExists := false
	for _, driver := range drivers {
		if driver == "sqlite3" {
			driverExists = true
		}
	}
	if !driverExists {
		sql.Register("sqlite3", &sqlite3.SQLiteDriver{})
	}
	return sql.Open("sqlite3", getDbFile())
}

func getDbFile() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	home := usr.HomeDir
	ambariManagerFolder := path.Join(home, ".ambarictl")
	if _, err := os.Stat(ambariManagerFolder); os.IsNotExist(err) {
		os.Mkdir(ambariManagerFolder, os.ModePerm)
	}
	return path.Join(ambariManagerFolder, "ambari-registry.db")
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
