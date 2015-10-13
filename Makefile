cross-build: linux darwin windows

linux: bin/carina-linux-amd64

darwin: bin/carina-darwin-amd64

windows: bin/carina.exe

bin/carina-linux-amd64: main.go
	 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/carina-linux-amd64 .

bin/carina-darwin-amd64: main.go
	 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/carina-darwin-amd64 .

bin/carina.exe: main.go
	 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/carina.exe .

.PHONY: clean

clean:
	 rm -f bin/*
