FROM golang:1.23rc1

ADD . /go/src/github.com/dokku/apps/web
WORKDIR /go/src/github.com/dokku/apps/web

RUN go install -v ./...
