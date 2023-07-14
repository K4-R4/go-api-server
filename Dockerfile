FROM golang:alpine

ENV GOPATH=

WORKDIR /go
ADD . /go
CMD ["go", "run", "main.go"]
