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
	"os/exec"
)

// DownloadViaScp downloads file from remote to local
func DownloadViaScp(sshConfig *easyssh.MakeConfig, source string, dest string) error {
	if len(sshConfig.Proxy.Server) > 0 {
		fmt.Print("TODO host jump!!!")
	}
	userAndRemote := fmt.Sprintf("%v@%v", sshConfig.User, sshConfig.Server)
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=no", "-q", "-P", sshConfig.Port, "-i", sshConfig.KeyPath, userAndRemote+":"+source, dest)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Copy %v (host: %v) to location: %v", source, sshConfig.Server, dest))
	return nil
}
