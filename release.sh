#!/bin/bash
set -euo pipefail

declare -xr ORG="getcarina"
declare -xr REPO="carina"
declare -xr BINARY=$REPO

function usage {
  echo 'usage: release.sh "v{major}.{minor}.{patch}" "Adjective Constellation"'
}

function main {
  # Pick your own leveled up tag
  TAG=${1:-}

  # Chosen from {adjective} {constellation}, where constellation comes from
  # http://www.astro.wisc.edu/~dolan/constellations/constellation_list.html
  NAME=${2:-}

  if [ "$TAG" == "" ] || [ "$NAME" == "" ] ; then
    usage
    exit 5
  fi

  echo "Releasing '$TAG' - $NAME: $DESCRIPTION"

  DESCRIPTION="Prototypal release of the Carina CLI"

  if [ ! -e "$( which github-release )" ]; then
    echo "You need github-release installed."
    echo "go get github.com/aktau/github-release"
    exit 1
  fi

  BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null)

  if [ "$BRANCH" != "master" ]; then
    echo "Must release from master branch"
    exit 2
  fi

  set +e
  git diff --exit-code > /dev/null
  if [ $? != 0 ]; then
    echo "Workspace is not clean. Exiting"
    exit 3
  fi
  set -e

  REMOTE="release"
  REMOTE_URL="git@github.com:${ORG}/${REPO}.git"

  #
  # Confirm that we have a remote named "Release"
  #

  set +e
  git remote show ${REMOTE} &> /dev/null

  rc=$?

  if [[ $rc != 0 ]]; then
    echo "Remote \"${REMOTE}\" not found. Exiting."
    exit 4
  fi
  set -e

  #
  # Now confirm that we've got the proper remote URL
  #

  REMOTE_ACTUAL_URL=$(git remote show release | grep Push | cut -d ":" -f2- | xargs)

  if [ "$REMOTE_URL" != "$REMOTE_ACTUAL_URL" ]; then
    echo -e "Remote \"${REMOTE}\" PUSH url incorrect.\nShould be ${REMOTE_URL}. Exiting."
    exit 5
  fi

  make clean
  # Build off master to make sure all is well
  make carina
  make test
  echo "Out with the old, in with the new"
  ./carina --version
  echo "---------------------------------"

  github-release release \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "$NAME" \
    --description "$DESCRIPTION"

  # Build with the tag now for actual binary shipping
  git fetch --tags release
  make build-tagged-for-release TAG=$TAG

  push_binaries
}

function push_binaries {
  github-release upload \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "${BINARY}-linux-amd64" \
    --file bin/${BINARY}-linux-amd64

  github-release upload \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "${BINARY}-darwin-amd64" \
    --file bin/${BINARY}-darwin-amd64

    github-release upload \
      --user "$ORG" \
      --repo "$REPO" \
      --tag "$TAG" \
      --name "${BINARY}.exe" \
      --file bin/${BINARY}.exe
}

main
