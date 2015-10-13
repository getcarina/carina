#!/bin/bash
set -euo pipefail

if [ ! -e "$( which github-release )" ]; then
  echo "You need github-release installed."
  echo "go get github.com/aktau/github-release"
  exit 2
fi

declare -xr ORG="carina"
declare -xr REPO="carina"
declare -xr BINARY=$REPO

TAG=${1}
NAME=${2}
DESCRIPTION="Prototypal release of the Carina CLI"

echo "Releasing '$TAG' - $NAME: $DESCRIPTION"

make cross-build

github-release release \
  --user "$ORG" \
  --repo "$REPO" \
  --tag "$TAG" \
  --pre-release \
  --name "$NAME" \
  --description "$DESCRIPTION"

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
