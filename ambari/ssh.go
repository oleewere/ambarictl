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
	"strconv"
	"sync"
	"time"
)

// RunAgentCommands executes bash commands on ambari agent hosts
func (a AmbariRegistry) RunAgentCommands(command string) {
	connectionProfileId := a.ConnectionProfile
	if len(connectionProfileId) == 0 {
		fmt.Println("No connection profile is attached for the active ambari server entry!")
		os.Exit(1)
	}
	connectionProfile := GetConnectionProfileById(connectionProfileId)

	agents := a.ListAgents()
	agentHosts := make([]string, len(agents))
	var wg sync.WaitGroup
	wg.Add(len(agents))
	for _, agent := range agents {
		agentHosts = append(agentHosts, agent.IP)
		ssh := &easyssh.MakeConfig{
			User:    connectionProfile.Username,
			Server:  agent.IP,
			KeyPath: connectionProfile.KeyPath,
			Port:    strconv.Itoa(connectionProfile.Port),
			Timeout: 60 * time.Second,
		}
		go func(ssh *easyssh.MakeConfig, command string, host string) {
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
			}
		}(ssh, command, agent.IP)
	}
	wg.Wait()
}
