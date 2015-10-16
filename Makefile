COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = rackerlabs
GITHUB_REPO = carina
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

build-in-docker: Dockerfile.build
	docker build -f Dockerfile.build -t carina-cli-build .
	docker run --rm carina-cli-build cat /builds.tgz | tar xz

tagged-build:
	git checkout $(TAG)
	make builds.tgz
	cp builds.tgz /builds.tgz

build-tagged-for-release:
	docker run --rm carina-cli-build sh -c "make --quiet tagged-build TAG=${TAG} && cat /builds.tgz" | tar xz

builds.tgz: cross-build
	tar -cvzf builds.tgz bin/*

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

.PHONY: clean build-in-docker build-tagged-for-release

clean:
	 rm -f bin/*
