# Ambari manager client
[![GoDoc Widget](https://godoc.org/github.com/oleewere/ambarictl/ambari?status.svg)](https://godoc.org/github.com/oleewere/ambarictl/ambari)
[![Build Status](https://travis-ci.org/oleewere/ambarictl.svg?branch=master)](https://travis-ci.org/oleewere/ambarictl)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleewere/ambarictl)](https://goreportcard.com/report/github.com/oleewere/ambarictl)
![license](http://img.shields.io/badge/license-Apache%20v2-blue.svg)

### Install

#### Installation on Mac OSX
```bash
brew tap oleewere/repo
brew install ambarictl
```

#### Installation on Linux

Using wget:
```bash
wget -qO- https://github.com/oleewere/ambarictl/releases/download/v0.3.0/ambarictl_0.3.0_linux_64-bit.tar.gz | tar -C /usr/bin -zxv ambarictl
```

Using curl:
```bash
curl -L -s https://github.com/oleewere/ambarictl/releases/download/v0.3.0/ambarictl_0.3.0_linux_64-bit.tar.gz | tar -C /usr/bin -xzv ambarictl
```

### Usage

#### Initialize Ambari registry db

```bash
ambarictl init
```

#### Create Ambari server entry
Ambari server entry contains informations about the Ambari server.
```bash
ambarictl create # it will ask inputs from the user like cluster name, Ambari server host etc.
```

#### Delete Ambari server entry
```bash
# use a Ambari server id that was created before
ambarictl delete $AMBARI_SERVER_ID
```

#### Create connection profile
Connection profile contains informations about how to ssh into Ambari agent machines.
```bash
ambarictl profiles create # it will ask inputs from the user like ssh key path, need host jump etc.
```

#### Attach connection profile to Ambari server
```bash
# use a profile id that was created before
ambarictl attach $CONNECTION_PROFILE_ID
```

#### Run example command on specific hosts
```bash
ambarictl run 'echo hello' -c INFRA_SOLR
```

#### Run example playbook
```bash
ambarictl playbook -f examples/print-configs.yml
```

#### Download logs for specific components
```bash
ambarictl logs -d /tmp/downloaded/logs -c INFRA_SOLR
```


### Developement
#### Build
```bash
make build
```
