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
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

// GetStringFlag trying to read a flag value, if it does not exists ask an input from the user
func GetStringFlag(flagValue string, defaultValue string, text string) string {
	if len(flagValue) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(text)
		if len(defaultValue) > 0 {
			fmt.Print(" (" + defaultValue + "): ")
		} else {
			fmt.Print(": ")
		}
		answer, _ := reader.ReadString('\n')
		if len(answer) == 0 || answer == "\n" {
			if len(defaultValue) == 0 {
				fmt.Println("Input cannot be empty!")
				os.Exit(1)
			}
			answer = defaultValue
		}
		return strings.TrimSpace(answer)
	}
	return flagValue
}

// GetStringFlag trying to read a password flag value, if it does not exists ask an input from the user
func GetPassword(flagValue string, text string) string {
	if len(flagValue) == 0 {
		fmt.Print(text + ": ")
		if terminal.IsTerminal(syscall.Stdin) {
			var fd = syscall.Stdin
			bytePassword, err := terminal.ReadPassword(fd)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			password := string(bytePassword)
			fmt.Println()
			return strings.TrimSpace(password)
		}
		answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if len(answer) == 0 || answer == "\n" {
			fmt.Println("Password cannot by empty")
			os.Exit(1)
		}
		return strings.TrimSpace(answer)

	}
	return flagValue
}
