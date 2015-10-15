COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = rackerlabs
GITHUB_REPO = carina
XFLAG_PRE = -X github.com/${GITHUB_ORG}/${GITHUB_REPO}
LDFLAGS = -w ${XFLAG_PRE}/version.Commit=${COMMIT} ${XFLAG_PRE}/version.Version=${VERSION}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = *.go version/*.go

carina: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o carina .

cross-build: carina linux darwin windows

linux: bin/carina-linux-amd64

darwin: bin/carina-darwin-amd64

windows: bin/carina.exe

bin/carina-linux-amd64: $(GOFILES)
	 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $@ .

bin/carina-darwin-amd64: $(GOFILES)
	 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $@ .

bin/carina.exe: $(GOFILES)
	 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $@ .

.PHONY: clean

clean:
	 rm -f bin/*
