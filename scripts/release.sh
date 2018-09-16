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

function print_help() {
  cat << EOF
   Usage: ./release.sh [additional options]
   --release-publish            run build and publish artifacts (on release tags)
   --release-major              create a tag with major version change (e.g.: 0.1.0 -> 1.0.0)
   --release-minor              create a tag with minor version change (e.g.: 0.1.0 -> 0.2.0)
   --release-patch              create a tag with patch version change (e.g.: 0.1.0 -> 0.1.1)
   -b, --release-build-only     create a dist from snapshot package
   -v, --version <version>      override ambari-python artifact versison
   -h, --help                   print help
EOF
}

function get_branch_name() {
  echo $(git rev-parse --abbrev-ref HEAD)
}

function get_last_release() {
  local last_rev=$(git rev-list --tags --max-count=1)
  if [[ -z "$last_rev" ]]; then
    echo "v0.0.0"
  else
    echo $(git describe --tags $last_rev)
  fi
}

function next_major_release() {
  echo $(get_version_number "$1") | awk '{split($0,a,"."); print "v"a[1]+1"."0"."0}'
}

function next_minor_release() {
  echo "$1" | awk '{split($0,a,"."); print a[1]"."a[2]+1"."0}'
}

function next_patch_release() {
  echo "$1" | awk '{split($0,a,"."); print a[1]"."a[2]"."a[3]+1}'
}

function new_branch_name() {
  echo "$1" | awk '{split($0,a,"."); print a[1]"."a[2]}'| sed -e 's/v/branch-/g'
}

function get_version_number() {
  local release_version="$1"
  echo "$release_version" | sed -e 's/v//g'
}

function get_head_tag() {
  echo $(git name-rev --tags --name-only $(git rev-parse HEAD))
}

function user_confirmation() {
  read -p "Continue (y/n)?" choice
  case "$choice" in
    y|Y|yes )
      echo "Release confirmed."
      ;;
    n|N|no )
      exit 0
      ;;
    * )
      echo "Invalid user anwser input"
      exit 1
      ;;
  esac
}

function release_and_push_new_branch() {
  local next_release="$1"
  local new_branch="$2"
  local version_number=$(get_version_number $next_release)
  git tag "$next_release"
  local release_result=$(run_release)
  if [[ "$release_result" == "0" ]]; then
    git push origin master
    git push origin $new_branch
  else
    echo "Pushing release tag was unsuccessful, revert tag creation. ('$next_release')"
    git tag -d $next_release
    exit 1
  fi
}

function release_and_push_actual_branch() {
  local next_release="$1"
  local actual_branch="$2"
  local version_number=$(get_version_number $next_release)
  git tag "$next_release" -m "$next_release (patch release)"
  local release_result=$(run_release)
  if [[ "$release_result" == "0" ]]; then
    git push origin $actual_branch
  else
    echo "Pushing release tag was unsuccessful, revert tag creation. ('$next_release')"
    git tag -d $next_release
    exit 1
  fi
}

function release_major() {
  echo "Create major release ..."
  local branch_name=$(get_branch_name)
  local last_release=$(get_last_release)
  echo "Branch name: $branch_name"
  if [[ "$branch_name" == "master" ]]; then
    echo "Last release: $last_release"
    local next_release=$(next_major_release $last_release)
    echo "New release: $next_release"
    local new_branch=$(new_branch_name $next_release)
    echo "New branch: $new_branch"
    user_confirmation
    release_and_push_new_branch "$next_release" "$new_branch"
  else
    echo "Major release can be created only on master branch. Exiting ..."
    exit 0
  fi
}

function release_minor() {
  echo "Create minor release ..."
  local branch_name=$(get_branch_name)
  local last_release=$(get_last_release)
  echo "Branch name: $branch_name"
  if [[ "$branch_name" == "master" ]]; then
    echo "Last release: $last_release"
    local next_release=$(next_minor_release $last_release)
    echo "New release: $next_release"
    local new_branch=$(new_branch_name $next_release)
    echo "New branch: $new_branch"
    user_confirmation
    release_and_push_new_branch "$next_release" "$new_branch"
  else
    echo "Minor release can be created only on master branch. Exiting ..."
    exit 0
  fi
}

function release_patch() {
  echo "Create patch release ..."
  local branch_name=$(get_branch_name)
  local last_release=$(get_last_release)
  echo "Branch name: $branch_name"
  if [[ "$branch_name" != "master" ]]; then
    if [[ "$branch_name" != branch* ]]; then
      echo "Cannot create patch release on feature branch"
      exit 0
    fi
    echo "Last release: $last_release"
    local next_release=$(next_patch_release $last_release)
    echo "New release: $next_release"
    user_confirmation
    release_and_push_actual_branch "$next_release" "$branch_name"
  else
    echo "Patch release cannot be created on master branch. Exiting ..."
    exit 0
  fi
}

function run_release() {
  docker run -w /go/src/github.com/oleewere/ambarictl -e GITHUB_TOKEN=$GITHUB_TOKEN --rm -v $AMBARICTL_ROOT_DIR/vendor/:/go/src/ -v $AMBARICTL_ROOT_DIR:/go/src/github.com/oleewere/ambarictl bepsays/ci-goreleaser:latest goreleaser --debug --rm-dist
}

function build_only() {
  docker run -w /go/src/github.com/oleewere/ambarictl -e GITHUB_TOKEN=$GITHUB_TOKEN --rm -v $AMBARICTL_ROOT_DIR/vendor/:/go/src/ -v $AMBARICTL_ROOT_DIR:/go/src/github.com/oleewere/ambarictl bepsays/ci-goreleaser:latest goreleaser --snapshot --debug --rm-dist --skip-publish
}

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
        --release-major)
          local RELEASE_MAJOR="true"
          shift 1
        ;;
        --release-minor)
          local RELEASE_MINOR="true"
          shift 1
        ;;
        --release-patch)
          local RELEASE_PATCH="true"
          shift 1
        ;;
        -h|--help)
          shift 1
          print_help
          exit 0
        ;;
        *)
          echo "Unknown option: $1"
          exit 1
        ;;
      esac
    done

  if [[ "$RELEASE_BUILD_ONLY" == "true" ]]; then
    build_only
  fi

  if [[ -z "$GITHUB_TOKEN" ]] ; then
    echo "Setting GITHUB_TOKEN variable is required."
    exit 1
  fi

  if [[ "$RELEASE_MAJOR" == "true" ]] ; then
    release_major
  fi

  if [[ "$RELEASE_MINOR" == "true" ]] ; then
    release_minor
  fi

  if [[ "$RELEASE_PATCH" == "true" ]] ; then
    release_patch
  fi
}


main ${1+"$@"}