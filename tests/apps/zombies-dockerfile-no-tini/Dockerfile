FROM golang:1.21.1

ADD . /go/src/github.com/dokku/apps/web
WORKDIR /go/src/github.com/dokku/apps/web

RUN go install -v ./...
