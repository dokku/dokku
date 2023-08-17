module github.com/dokku/dokko/tests/apps/gorpc

go 1.19

require (
	github.com/golang/protobuf v1.5.3
	golang.org/x/net v0.14.0
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/grpc v1.29.1
)

require (
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)

replace google.golang.org/grpc/examples/helloworld/helloworld => ./helloworld
