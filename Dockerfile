FROM golang:1.10

WORKDIR $GOPATH/src/github.com/areller/podlog
ADD . .

WORKDIR $GOPATH/src/github.com/areller/podlog/cmd/podlog

RUN go get
RUN go build

CMD ["./podlog"]