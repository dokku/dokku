FROM golang:1.21.6

ADD . /go/src/github.com/dokku/apps/web
WORKDIR /go/src/github.com/dokku/apps/web

RUN go install -v ./...
