SHELL := /bin/bash

COMMIT = $(shell git rev-parse --verify HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
PERMALINK = $(shell if [[ $(VERSION) =~ [^-]*-([^.]+).* ]]; then echo $${BASH_REMATCH[1]}; else echo "latest"; fi)

GITHUB_ORG = getcarina
GITHUB_REPO = carina
REPO_PATH = $(GOPATH)/src/github.com/$(GITHUB_ORG)/$(GITHUB_REPO)

XFLAG_PRE = -X github.com/$(GITHUB_ORG)/$(GITHUB_REPO)
LDFLAGS = -w $(XFLAG_PRE)/version.Commit=$(COMMIT) $(XFLAG_PRE)/version.Version=$(VERSION)

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = $(wildcard **/*.go)
GOFILES_NOVENDOR = $(shell go list ./... | grep -v /vendor/)

BINDIR = bin/carina/$(VERSION)

default: get-deps validate local

get-deps:
	@#go get github.com/golang/lint/golint
	script/install-glide.sh
	glide install --force --update-vendored

validate:
	go fmt $(GOFILES_NOVENDOR)
	go vet $(GOFILES_NOVENDOR)
	@#go list ./... | grep -v /vendor/ | xargs -L1 golint --set_exit_status

local: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o carina .

test: local
	go test $(GOFILES_NOVENDOR)
	eval "$( ./carina bash-completion )"
	./carina --version

cross-build: linux darwin windows
	cp -R $(BINDIR) bin/carina/${PERMALINK}

linux: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Linux/x86_64/carina .
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINDIR)/Linux/i686/carina .

darwin: $(GOFILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Darwin/x86_64/carina .

windows: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Windows/x86_64/carina.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(BINDIR)/Windows/i686/carina.exe .

.PHONY: clean deploy

clean:
	-rm -fr vendor
	-rm -fr bin
	-rm carina

deploy:
	curl -O https://ec4a542dbf90c03b9f75-b342aba65414ad802720b41e8159cf45.ssl.cf5.rackcdn.com/1.2/Linux/amd64/rack
	chmod +x rack
	./rack files object upload-dir --recurse --container carina-downloads --dir bin
	curl -X POST -d '' https://www.myget.org/BuildSource/Hook/rackspace?identifier=0f9b4abd-6ec9-4b29-aab5-541967d4d260
