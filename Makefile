COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = rackerlabs
GITHUB_REPO = carina

REPO_PATH = ${GOPATH}/src/github.com/${GITHUB_ORG}/${GITHUB_REPO}

XFLAG_PRE = -X github.com/${GITHUB_ORG}/${GITHUB_REPO}
LDFLAGS = -w ${XFLAG_PRE}/version.Commit=${COMMIT} ${XFLAG_PRE}/version.Version=${VERSION}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = *.go version/*.go

default: carina

get-deps:
	go get ./...

carina: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o carina .

gocarina: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o ${GOPATH}/bin/carina .

cross-build: get-deps carina linux darwin windows

build-tagged-for-release: clean
	-docker rm -fv carina-build
	docker build -f Dockerfile.build -t carina-cli-build --no-cache=true .
	docker run --name carina-build carina-cli-build make tagged-build TAG=$(TAG)
	mkdir -p bin/
	docker cp carina-build:/built/bin .

checkout-tag:
	git checkout $(TAG)

# This one is intended to be run inside the accompanying Docker container
tagged-build: checkout-tag cross-build
	./carina --version
	mkdir -p /built/
	cp -r bin /built/bin

linux: bin/carina-linux-amd64

darwin: bin/carina-darwin-amd64

windows: bin/carina.exe

bin/carina-linux-amd64: $(GOFILES)
	 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $@ .

bin/carina-darwin-amd64: $(GOFILES)
	 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $@ .

bin/carina.exe: $(GOFILES)
	 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $@ .

test: carina
	@echo "Tests are cool, we should do those."
	./carina --version

.PHONY: clean build-tagged-for-release checkout tagged-build

clean:
	 rm -f bin/*
