FROM golang:1.22.6

RUN mkdir /app
WORKDIR /app

ADD . /app
RUN go install ./...

CMD /go/bin/greeter_server
