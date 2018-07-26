package ambari

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"strconv"
	"os"
	"os/user"
	"path"
)

func CreateAmbariRegistryDb() {
	db, err := sql.Open("sqlite3", GetDbFile())
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS ambari_registry (id VARCHAR PRIMARY KEY, hostname VARCHAR, port INTEGER, protocol VARCHAR, username VARCHAR, password VARCHAR, cluster TEXT, active INTEGER)")
	checkErr(err)
	statement.Exec()
}

func DropAmbariRegistryRecords() {
	db, err := sql.Open("sqlite3", GetDbFile())
	checkErr(err)
	defer db.Close()
	statement, err := db.Prepare("DELETE from ambari_registry")
	checkErr(err)
	statement.Exec()
}

func ListAmbariRegistryEntries() {
	db, err := sql.Open("sqlite3", GetDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id,hostname,port,protocol,username,cluster,active FROM ambari_registry")
	checkErr(err)
	var id string
	var hostname string
	var port int
	var protocol string
	var username string
	var cluster string
	var active int
	for rows.Next() {
		rows.Scan(&id, &hostname, &port, &protocol, &username, &cluster, &active)
		activeValue := false
		if active == 1 {
			activeValue = true
		}
		rowDetails := fmt.Sprintf("%s - %s://%s:%v - %s - %s / ******** - active: %v", id, protocol, hostname, port, cluster, username, activeValue)
		fmt.Println(rowDetails)
	}
	rows.Close()
}

func RegisterNewAmbariEntry(id string, hostname string, port int, protocol string, username string, password string, cluster string) {
	db, err := sql.Open("sqlite3", GetDbFile())
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT id FROM ambari_registry WHERE id = '" + id + "'")
	checkErr(err)
	var check_id string
	for rows.Next() {
		rows.Scan(&check_id)
	}
	rows.Close()
	if len(check_id) > 0 {
		alreadyExistMsg := fmt.Sprintf("Registry with id '%s' is already defined as a registry entry", check_id)
		fmt.Println(alreadyExistMsg)
		os.Exit(1)
	}

	statement, _ := db.Prepare("INSERT INTO ambari_registry (id, hostname, port, protocol, username, password, cluster, active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	_, insertErr := statement.Exec(id, hostname, strconv.Itoa(port), protocol, username, password, cluster, strconv.Itoa(1))
	checkErr(insertErr)
}

func GetActiveAmbari() AmbariRegistry {
	db, err := sql.Open("sqlite3", GetDbFile())
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

	return AmbariRegistry{name: id, hostname:hostname, port: port, protocol: protocol, username: username, password: password, cluster: cluster, active: 1}
}

func GetDbFile() string {
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