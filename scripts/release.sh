#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

readlinkf(){
  # get real path on mac OSX
  perl -MCwd -e 'print Cwd::abs_path shift' "$1";
}

if [ "$(uname -s)" = 'Linux' ]; then
  AMBARICTL_SCRIPT_DIR="`dirname "$(readlink -f "$0")"`"
else
  AMBARICTL_SCRIPT_DIR="`dirname "$(readlinkf "$0")"`"
fi

AMBARICTL_ROOT_DIR="`dirname \"$AMBARICTL_SCRIPT_DIR\"`"

function main() {

  local RELEASE_BUILD_ONLY="false"
  local RELEASE="false"

  while [[ $# -gt 0 ]]
    do
      key="$1"
      case $key in
        -b|--release-build-only)
          shift 1
          RELEASE_BUILD_ONLY="true"
        ;;
        -r|--release)
          shift 1
          RELEASE="true"
        ;;
        *)
          echo "Unknown option: $1"
          exit 1
        ;;
      esac
    done

  if [[ -z "$GITHUB_TOKEN" ]] ; then
    echo "Setting GITHUB_TOKEN variable is required."
    exit 1
  fi

  if [[ "$RELEASE" == "true" ]]; then
    echo "Release ..."
    exit 0
  elif [[ "$RELEASE_BUILD_ONLY" == "true" ]]; then
    docker run -w /go/src/github.com/oleewere/ambarictl -e GITHUB_TOKEN=$GITHUB_TOKEN --rm -v $AMBARICTL_ROOT_DIR/vendor/:/go/src/ -v $AMBARICTL_ROOT_DIR:/go/src/github.com/oleewere/ambarictl bepsays/ci-goreleaser:latest goreleaser --snapshot --debug --rm-dist --skip-publish
  fi
}


main ${1+"$@"}