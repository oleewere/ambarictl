# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GIT_REV_SHORT = $(shell git rev-parse --short HEAD)
GITHUB_TOKEN := $(shell git config --global --get github.token || echo $$GITHUB_TOKEN)

ifeq ("$(shell git rev-list --tags --max-count=1)", "")
  LAST_RELEASE="v0.0.0"
else
  LAST_RELEASE=$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
endif

SNAPSHOT_VERSION=$(shell echo $(LAST_RELEASE) | awk '{split($$0,a,"."); print "v"a[1]+1"."0"."0}')

ifeq ("$(shell git name-rev --tags --name-only $(shell git rev-parse HEAD))", "undefined")
	VERSION_FOR_BUILD="$(SNAPSHOT_VERSION)-SNAPSHOT"
else
	VERSION_FOR_BUILD=$(shell git name-rev --tags --name-only $(shell git rev-parse HEAD) | sed 's/\^.*$///')
endif

install:
	go install -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION_FOR_BUILD)" .

build:
	go build -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION_FOR_BUILD)" -o ambarictl .

test:
	go test

all: build test

clean:
	rm -rf dist

last-release:
	@echo "Last release version: $(LAST_RELEASE)"

version:
	@echo "Release/Snapshot version: $(VERSION_FOR_BUILD)"

binary:
	./scripts/release.sh --release-build-only

major-release:
	./scripts/release.sh --release-major

minor-release:
	./scripts/release.sh --release-minor

patch-release:
	./scripts/release.sh --release-patch
