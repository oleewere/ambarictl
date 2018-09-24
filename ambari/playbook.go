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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Playbook contains an array of tasks that will be executed on ambari hosts
type Playbook struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tasks       []Task `yaml:"tasks"`
}

// Task represent a task that can be executed on an ambari hosts
type Task struct {
	Name                string `yaml:"name"`
	Command             string `yaml:"command"`
	Download            string `yaml:"download"`
	Copy                string `yaml:"copy"`
	LocalCopy           string `yaml:"local_copy"`
	LocalCommand        string `yaml:"local_command"`
	AmbariCommand       string `yaml:"ambari_command"`
	HostComponentFilter string `yaml:"host_component_filter"`
	AmbariServerFilter  bool   `yaml:"ambari_server"`
	AmbariAgentFilter   bool   `yaml:"ambari_agent"`
	HostFilter          string `yaml:"hosts"`
	ServiceFilter       string `yaml:"services"`
	ComponentFilter     string `yaml:"components"`
}

// LoadPlaybookFile read a playbook yaml file and transform it to a Playbook object
func LoadPlaybookFile(location string) Playbook {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	playbook := Playbook{}
	err = yaml.Unmarshal([]byte(data), &playbook)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("[Executing playbook: %v, file: %v]", playbook.Name, location))
	return playbook
}

// ExecutePlaybook runs tasks on ambari hosts based on a playbook object
func (a AmbariRegistry) ExecutePlaybook(playbook Playbook) {
	tasks := playbook.Tasks
	for _, task := range tasks {
		if len(task.Command) > 0 {
			if task.AmbariAgentFilter {
				a.RunRemoteHostCommand(task.Command, nil)
			} else {
				filter := CreateFilter(task.ServiceFilter, task.ComponentFilter, task.HostFilter, task.AmbariServerFilter)
				a.RunRemoteHostCommand(task.Command, a.GetFilteredHosts(filter))
			}
		}
	}
}
