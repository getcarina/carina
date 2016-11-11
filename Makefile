COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = getcarina
GITHUB_REPO = carina

REPO_PATH = ${GOPATH}/src/github.com/${GITHUB_ORG}/${GITHUB_REPO}

XFLAG_PRE = -X github.com/${GITHUB_ORG}/${GITHUB_REPO}
LDFLAGS = -w ${XFLAG_PRE}/version.Commit=${COMMIT} ${XFLAG_PRE}/version.Version=${VERSION}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = $(wildcard **/*.go)
GOFILES_NOVENDOR = $(shell go list ./... | grep -v /vendor/)

BINDIR = bin/carina/$(VERSION)

default: get-deps validate local

get-deps:
	go get github.com/golang/lint/golint
	go get github.com/Masterminds/glide
	glide install --force --update-vendored

validate:
	go fmt $(GOFILES_NOVENDOR)
	go vet $(GOFILES_NOVENDOR)
	go list ./... | grep -v /vendor/ | xargs -L1 golint --set_exit_status

local: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o carina .

test: local
	go test -v $(GOFILES_NOVENDOR)
	eval "$( ./carina bash-completion )"
	./carina --version

carina-linux: linux
	cp bin/carina-linux-amd64 carina-linux

cross-build: linux darwin windows
	cp -R $(BINDIR) bin/carina/latest

linux: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Linux/x86_64/carina .
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINDIR)/Linux/i686/carina .

darwin: $(GOFILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Darwin/x86_64/carina .

windows: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Windows/x86_64/carina.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(BINDIR)/Windows/i686/carina.exe .

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
	-rm -fr vendor
	-rm -fr bin
	-rm carina
	-rm carina-linux
	-rm ca-certificates.crt
