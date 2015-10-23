FROM golang:1.5

ADD . $GOPATH/src/github.com/getcarina/carina/
WORKDIR $GOPATH/src/github.com/getcarina/carina/

RUN make get-deps
