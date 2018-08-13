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
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"
)

// NOTE: majority of the code copied from easy-ssh proxy see: github.com/appleboy/easyssh-proxy
type (
	// MakeConfig Contains main authority information.
	// User field should be a name of user on remote server (ex. john in ssh john@example.com).
	// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
	// Key is a path to private key on your local machine.
	// Port is SSH server port on remote machine.
	// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
	// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
	MakeConfig struct {
		User     string
		Server   string
		Key      string
		KeyPath  string
		Port     string
		Password string
		Timeout  time.Duration
		Proxy    DefaultConfig
	}

	// DefaultConfig for ssh proxy config
	DefaultConfig struct {
		User     string
		Server   string
		Key      string
		KeyPath  string
		Port     string
		Password string
		Timeout  time.Duration
	}
)

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	pubkey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

func getSSHConfig(config DefaultConfig) *ssh.ClientConfig {
	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if config.Password != "" {
		auths = append(auths, ssh.Password(config.Password))
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer sshAgent.Close()
	}

	if config.KeyPath != "" {
		if pubkey, err := getKeyFile(config.KeyPath); err == nil {
			auths = append(auths, ssh.PublicKeys(pubkey))
		}
	}

	if config.Key != "" {
		if signer, err := ssh.ParsePrivateKey([]byte(config.Key)); err == nil {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	}

	return &ssh.ClientConfig{
		Timeout:         config.Timeout,
		User:            config.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// connect to remote server using MakeConfig struct and returns *ssh.Session
func (sshConf *MakeConfig) connect() (*ssh.Session, error) {
	var client *ssh.Client
	var err error

	targetConfig := getSSHConfig(DefaultConfig{
		User:     sshConf.User,
		Key:      sshConf.Key,
		KeyPath:  sshConf.KeyPath,
		Password: sshConf.Password,
		Timeout:  sshConf.Timeout,
	})

	// Enable proxy command
	if sshConf.Proxy.Server != "" {
		proxyConfig := getSSHConfig(DefaultConfig{
			User:     sshConf.Proxy.User,
			Key:      sshConf.Proxy.Key,
			KeyPath:  sshConf.Proxy.KeyPath,
			Password: sshConf.Proxy.Password,
			Timeout:  sshConf.Proxy.Timeout,
		})

		proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(sshConf.Proxy.Server, sshConf.Proxy.Port), proxyConfig)
		if err != nil {
			return nil, err
		}

		conn, err := proxyClient.Dial("tcp", net.JoinHostPort(sshConf.Server, sshConf.Port))
		if err != nil {
			return nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(sshConf.Server, sshConf.Port), targetConfig)
		if err != nil {
			return nil, err
		}

		client = ssh.NewClient(ncc, chans, reqs)
	} else {
		client, err = ssh.Dial("tcp", net.JoinHostPort(sshConf.Server, sshConf.Port), targetConfig)
		if err != nil {
			return nil, err
		}
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (sshConf *MakeConfig) Stream(command string, timeout int) (<-chan string, <-chan string, <-chan bool, <-chan error, error) {
	// continuously send the command's output over the channel
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	doneChan := make(chan bool)
	errChan := make(chan error)

	// connect to remote host
	session, err := sshConf.connect()
	if err != nil {
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	// defer session.Close()
	// connect to both outputs (they are of type io.Reader)
	outReader, err := session.StdoutPipe()
	if err != nil {
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	errReader, err := session.StderrPipe()
	if err != nil {
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	err = session.Start(command)
	if err != nil {
		return stdoutChan, stderrChan, doneChan, errChan, err
	}

	// combine outputs, create a line-by-line scanner
	stdoutReader := io.MultiReader(outReader)
	stderrReader := io.MultiReader(errReader)
	stdoutScanner := bufio.NewScanner(stdoutReader)
	stderrScanner := bufio.NewScanner(stderrReader)

	go func(stdoutScanner, stderrScanner *bufio.Scanner, stdoutChan, stderrChan chan string, doneChan chan bool, errChan chan error) {
		defer close(stdoutChan)
		defer close(stderrChan)
		defer close(doneChan)
		defer close(errChan)
		defer session.Close()

		timeoutChan := time.After(time.Duration(timeout) * time.Second)
		res := make(chan bool, 1)

		go func() {
			for stdoutScanner.Scan() {
				stdoutChan <- stdoutScanner.Text()
			}
			for stderrScanner.Scan() {
				stderrChan <- stderrScanner.Text()
			}
			// close all of our open resources
			res <- true
		}()

		select {
		case <-res:
			errChan <- session.Wait()
			doneChan <- true
		case <-timeoutChan:
			stderrChan <- "Run Command Timeout!"
			errChan <- nil
			doneChan <- false
		}
	}(stdoutScanner, stderrScanner, stdoutChan, stderrChan, doneChan, errChan)

	return stdoutChan, stderrChan, doneChan, errChan, err
}

// Run command on remote machine and returns its stdout as a string
func (sshConf *MakeConfig) Run(command string, timeout int) (outStr string, errStr string, isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, errChan, err := sshConf.Stream(command, timeout)
	if err != nil {
		return outStr, errStr, isTimeout, err
	}
	// read from the output channel until the done signal is passed
loop:
	for {
		select {
		case isTimeout = <-doneChan:
			break loop
		case outline := <-stdoutChan:
			if outline != "" {
				outStr += outline + "\n"
			}
		case errline := <-stderrChan:
			if errline != "" {
				errStr += errline + "\n"
			}
		case err = <-errChan:
		}
	}
	// return the concatenation of all signals from the output channel
	return outStr, errStr, isTimeout, err
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (sshConf *MakeConfig) Scp(sourceFile string, etargetFile string) error {
	session, err := sshConf.connect()

	if err != nil {
		return err
	}
	defer session.Close()

	targetFile := filepath.Base(etargetFile)

	src, srcErr := os.Open(sourceFile)

	if srcErr != nil {
		return srcErr
	}

	srcStat, statErr := src.Stat()

	if statErr != nil {
		return statErr
	}

	go func() {
		w, err := session.StdinPipe()

		if err != nil {
			return
		}
		defer w.Close()

		fmt.Fprintln(w, "C0644", srcStat.Size(), targetFile)

		if srcStat.Size() > 0 {
			io.Copy(w, src)
			fmt.Fprint(w, "\x00")
		} else {
			fmt.Fprint(w, "\x00")
		}
	}()

	runErr := session.Run(fmt.Sprintf("scp -tr %s", etargetFile));
	if runErr != nil {
		return runErr
	}

	return nil
}
