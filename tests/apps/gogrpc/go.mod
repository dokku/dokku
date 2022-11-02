module github.com/dokku/dokko/tests/apps/gorpc

go 1.19

require (
	github.com/golang/protobuf v1.5.2
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/grpc v1.29.1
)

require (
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)

replace google.golang.org/grpc/examples/helloworld/helloworld => ./helloworld
