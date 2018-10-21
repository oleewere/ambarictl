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
	"fmt"
	"github.com/appleboy/easyssh-proxy"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

// RemoteResponse represents an ssh command output
type RemoteResponse struct {
	StdOut string
	StdErr string
	Done   bool
}

// RunRemoteHostCommand executes bash commands on ambari agent hosts
func (a AmbariRegistry) RunRemoteHostCommand(command string, filteredHosts map[string]bool) map[string]RemoteResponse {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)
	var hosts map[string]bool
	if len(filteredHosts) > 0 {
		hosts = filteredHosts
	} else {
		hosts = a.GetFilteredHosts(Filter{})
	}
	response := make(map[string]RemoteResponse)
	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for host := range hosts {
		ssh := createSshConfig(connectionProfile, host)
		go func(ssh *easyssh.MakeConfig, command string, host string, response map[string]RemoteResponse) {
			defer wg.Done()
			stdout, stderr, done, err := ssh.Run(command, 60)
			// Handle errors
			msgHeader := fmt.Sprintf("%v (done: %v) - output:", host, done)
			fmt.Println(msgHeader)
			if err != nil {
				panic("Can't run remote command: " + err.Error())
			} else {
				if len(stdout) > 0 {
					fmt.Println(stdout)
				}
				if len(stderr) > 0 {
					fmt.Println("std error:")
					fmt.Println(stderr)
				}
				response[host] = RemoteResponse{StdOut: stdout, StdErr: stderr, Done: done}
			}
		}(ssh, command, host, response)
	}
	wg.Wait()
	return response
}

// CopyToRemote copy local file to remote host(s)
func (a AmbariRegistry) CopyToRemote(source string, dest string, filteredHosts map[string]bool) {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)
	var hosts map[string]bool
	if len(filteredHosts) > 0 {
		hosts = filteredHosts
	} else {
		hosts = a.GetFilteredHosts(Filter{})
	}
	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for host := range hosts {
		ssh := createSshConfig(connectionProfile, host)
		go func(ssh *easyssh.MakeConfig, source string, dest string, host string) {
			defer wg.Done()
			err := ssh.Scp(source, dest)
			// Handle errors
			if err != nil {
				errMsg := fmt.Sprintf("Can't run remote command on host '%v (scp %v to %v)", host, source, dest)
				fmt.Println(errMsg)
			} else {
				succMsg := fmt.Sprintf("Copying to remote host '%v' is successful. (from - %v, to %v)", host, source, dest)
				fmt.Println(succMsg)
			}
		}(ssh, source, dest, host)
	}
	wg.Wait()
}

// CopyFromRemote copy 1 file from 1 remote host to locally
func (a AmbariRegistry) CopyFromRemote(source string, dest string, host string) {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)
	ssh := createSshConfig(connectionProfile, host)
	err := DownloadViaScp(ssh, source, dest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// CopyFromRemoteHosts copy remote file to remote host(s)
func (a AmbariRegistry) CopyFromRemoteHosts(source string, dest string, filteredHosts map[string]bool) {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)
	var hosts map[string]bool
	if len(filteredHosts) > 0 {
		hosts = filteredHosts
	} else {
		hosts = a.GetFilteredHosts(Filter{})
	}
	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for host := range hosts {
		ssh := createSshConfig(connectionProfile, host)
		go func(ssh *easyssh.MakeConfig, source string, dest string, host string) {
			defer wg.Done()
			hostFolder := path.Join(dest, host)
			os.MkdirAll(hostFolder, os.ModePerm)
			err := DownloadViaScp(ssh, source, hostFolder)
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to copy from host '%v', reason:", host))
				fmt.Println(err)
			}
		}(ssh, source, dest, host)
	}
	wg.Wait()
}

// CopyFolderFromRemote copy folder (zipping it first) to local filesystem from remote location
func (a AmbariRegistry) CopyFolderFromRemote(component string, source string, dest string, filteredHosts map[string]bool) {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)
	var hosts map[string]bool
	if len(filteredHosts) > 0 {
		hosts = filteredHosts
	} else {
		hosts = a.GetFilteredHosts(Filter{})
	}

	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for host := range hosts {
		ssh := createSshConfig(connectionProfile, host)
		go func(ssh *easyssh.MakeConfig, component string, source string, dest string, host string) {
			defer wg.Done()
			tmpSource := fmt.Sprintf("/tmp/%v.tar.gz", component)
			command := fmt.Sprintf("cd %v && tar -cvf %v *", source, tmpSource)
			stdout, stderr, _, err := ssh.Run(command, 60)
			// Handle errors
			if err != nil {
				panic("Can't run remote command: " + err.Error())
			} else {
				if len(stdout) > 0 {
					fmt.Println(fmt.Sprintf("Zipping '%v' log files has been finished on host %v", component, host))
				}
				if len(stderr) > 0 {
					fmt.Println("std error:")
					fmt.Println(stderr)
				}
			}
			hostFolder := path.Join(dest, host)
			os.MkdirAll(hostFolder, os.ModePerm)
			err = DownloadViaScp(ssh, tmpSource, hostFolder)
			if err != nil {
				fmt.Println(err)
			}
		}(ssh, component, source, dest, host)
	}
	wg.Wait()
}

func createSshConfig(connectionProfile ConnectionProfile, host string) *easyssh.MakeConfig {
	if len(connectionProfile.ProxyAddress) > 0 {
		return &easyssh.MakeConfig{
			User:    connectionProfile.Username,
			Server:  host,
			KeyPath: connectionProfile.KeyPath,
			Port:    strconv.Itoa(connectionProfile.Port),
			Timeout: 60 * time.Second,
			Proxy: easyssh.DefaultConfig{
				User:    connectionProfile.Username,
				Server:  connectionProfile.ProxyAddress,
				Port:    strconv.Itoa(connectionProfile.Port),
				KeyPath: connectionProfile.KeyPath,
				Timeout: 60 * time.Second,
			},
		}
	}
	return &easyssh.MakeConfig{
		User:    connectionProfile.Username,
		Server:  host,
		KeyPath: connectionProfile.KeyPath,
		Port:    strconv.Itoa(connectionProfile.Port),
		Timeout: 60 * time.Second,
	}
}
