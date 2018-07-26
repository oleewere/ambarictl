VERSION = 1.0.0
GIT_REV_SHORT = $(shell git rev-parse --short HEAD)

install:
	go install

build:
	go build -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION)" .

test:
	go test

all: build test