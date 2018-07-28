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

// CreateAmbariRegistryDb initialize ambari-manager database
func CreateAmbariRegistryDb() {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS ambari_registry (id VARCHAR PRIMARY KEY, hostname VARCHAR, port INTEGER, protocol VARCHAR, username VARCHAR, password VARCHAR, cluster TEXT, active INTEGER)")
	checkErr(err)
	statement.Exec()
}

// DropAmbariRegistryRecords drop all entries from ambari-manager database
func DropAmbariRegistryRecords() {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("DELETE from ambari_registry")
	checkErr(err)
	statement.Exec()
}

// ListAmbariRegistryEntries get all ambari registries from ambari-manager database
func ListAmbariRegistryEntries() []AmbariRegistry {
	db, err := getDb()
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,password,cluster,active FROM ambari_registry")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var password string
	var cluster string
	var active int
	var ambariRegistries []AmbariRegistry
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &password, &cluster, &active)
		ambariRegistry := AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol,
			Username: username, Password: password, Cluster: cluster, Active: active}
		ambariRegistries = append(ambariRegistries, ambariRegistry)
	}
	rows.Close()
	return ambariRegistries
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

// RegisterNewAmbariEntry create new ambari registry entry in ambari-manager database
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

// GetActiveAmbari get the active ambari registry from ambari-manager database (should be only one)
func GetActiveAmbari() AmbariRegistry {
	db, err := sql.Open("sqlite3", getDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,password,cluster FROM ambari_registry WHERE active = '1'")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var password string
	var cluster string
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &password, &cluster)
	}
	rows.Close()

	return AmbariRegistry{Name: id, Hostname: hostname, Port: port, Protocol: protocol, Username: username, Password: password, Cluster: cluster, Active: 1}
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
	ambariManagerFolder := path.Join(home, ".ambari-manager")
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
