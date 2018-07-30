VERSION = 1.0.0
GIT_REV_SHORT = $(shell git rev-parse --short HEAD)

build:
	go build -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION)" -o ambarictl .

install:
	go install

test:
	go test

all: build test
