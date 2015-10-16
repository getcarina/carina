#!/bin/bash
set -euo pipefail

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

git diff --exit-code > /dev/null
if [ $? != 0 ]; then
  echo "Workspace is not clean. Exiting"
  exit 3
fi

declare -xr ORG="rackerlabs"
declare -xr REPO="carina"
declare -xr BINARY=$REPO

# Pick your own leveled up tag
TAG=${1}

# Chosen from {adjective} {constellation}, where constellation comes from
# http://www.astro.wisc.edu/~dolan/constellations/constellation_list.html
NAME=${2}

DESCRIPTION="Prototypal release of the Carina CLI"

echo "Releasing '$TAG' - $NAME: $DESCRIPTION"

make clean
# Build off master to make sure all is well
make build-in-docker

github-release release \
  --user "$ORG" \
  --repo "$REPO" \
  --tag "$TAG" \
  --pre-release \
  --name "$NAME" \
  --description "$DESCRIPTION"

# Build with the tag now for actual binary shipping
git pull release master
git checkout "$TAG"
make build-tagged-for-release TAG=$TAG

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

git checkout master
