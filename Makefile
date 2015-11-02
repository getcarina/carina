COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = getcarina
GITHUB_REPO = carina

REPO_PATH = ${GOPATH}/src/github.com/${GITHUB_ORG}/${GITHUB_REPO}

XFLAG_PRE = -X github.com/${GITHUB_ORG}/${GITHUB_REPO}
LDFLAGS = -w ${XFLAG_PRE}/version.Commit=${COMMIT} ${XFLAG_PRE}/version.Version=${VERSION}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = main.go version/*.go

default: carina

get-deps:
	go get ./...

carina-linux: linux
	cp bin/carina-linux-amd64 carina-linux

test: carina
	go test -v
	eval "$( ./carina --bash-completion )"
	./carina --version

gocarina: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o ${GOPATH}/bin/carina .

cross-build: get-deps carina linux darwin windows

carina: get-deps $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o carina .

linux: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/carina-linux-amd64 .

darwin: $(GOFILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/carina-darwin-amd64 .

windows: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/carina.exe .



############################ RELEASE TARGETS ############################

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

############################## DOCKER IMAGE ###############################

ca-certificates.crt:
	-docker rm -fv carina-cert-grab
	docker run --name carina-cert-grab ubuntu:15.04 sh -c "apt-get update -y && apt-get install ca-certificates -y"
	docker cp carina-cert-grab:/etc/ssl/certs/ca-certificates.crt .

carina/cli: ca-certificates.crt carina-linux
	docker build -t carina/cli .

.PHONY: clean build-tagged-for-release checkout tagged-build

clean:
	 -rm -f bin/*
	 -rm carina
	 -rm carina-linux
	 -rm ca-certificates.crt
