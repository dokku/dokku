module github.com/dokku/dokko/tests/apps/gorpc

go 1.21

require (
	github.com/golang/protobuf v1.5.3
	golang.org/x/net v0.20.0
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/grpc v1.60.1
)

require (
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace google.golang.org/grpc/examples/helloworld/helloworld => ./helloworld
