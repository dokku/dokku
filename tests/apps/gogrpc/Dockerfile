FROM golang:1.24.3

RUN mkdir /app
WORKDIR /app

ADD . /app
RUN go install ./...

CMD /go/bin/greeter_server
