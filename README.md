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
wget -qO- https://github.com/oleewere/ambarictl/releases/download/v0.1.1/ambarictl_0.1.1_linux_64-bit.tar.gz | tar -C /usr/bin -zxv ambarictl
```

Using curl:
```bash
curl -L -s https://github.com/oleewere/ambarictl/releases/download/v0.1.1/ambarictl_0.1.1_linux_64-bit.tar.gz | tar -C /usr/bin -xzv ambarictl
```

### Developement
#### Build
```bash
make build
```
