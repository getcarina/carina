#!/usr/bin/env bash
set -euo pipefail

if type glide &> /dev/null ; then
    version=`glide --version`
    if [[ "$version" =~ "0.11.1"  ]]; then
        echo "Found glide v0.11.1"
        exit 0
    else
        echo "You have glide installed but it is the wrong version($version). You can either remove glide, and re-build carina to have it installed automatically for you, or you must manually get the glide binary on your path to v0.11.1."
        exit 1
    fi
fi

echo "Installing glide v0.11.1..."
if [ ! -d "$GOPATH/src/github.com/Masterminds/glide" ]; then
  git clone https://github.com/Masterminds/glide.git $GOPATH/src/github.com/Masterminds/glide
fi
cd $GOPATH/src/github.com/Masterminds/glide
git checkout v0.11.1
go install --ldflags="-X main.version=v0.11.1" github.com/Masterminds/glide

if type glide &> /dev/null ; then
    echo "Done!"
    exit 0
else
    echo "Could not find glide after installing it. Make sure GOPATH/bin is in PATH. Aborting..."
    exit 1
fi
