VERSION = 0.1.0
GIT_REV_SHORT = $(shell git rev-parse --short HEAD)
GITHUB_TOKEN := $(shell git config --global --get github.token || echo $$GITHUB_TOKEN)

install:
	go install -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION)" .

build:
	go build -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION)" -o ambarictl .

test:
	go test

all: build test

clean:
	rm -rf dist

binary:
	./scripts/release.sh --release-build-only

release:
	./scripts/release.sh --release