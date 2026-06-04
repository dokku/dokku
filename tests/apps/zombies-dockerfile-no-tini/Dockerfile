FROM golang:1.26.4

ADD . /go/src/github.com/dokku/apps/web
WORKDIR /go/src/github.com/dokku/apps/web

RUN go install -v ./...
