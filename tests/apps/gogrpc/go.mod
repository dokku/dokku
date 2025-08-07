module github.com/dokku/dokko/tests/apps/gorpc

go 1.23.2

toolchain go1.24.1

require (
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/grpc v1.74.2
)

require (
	google.golang.org/grpc/examples v0.0.0-20240118175532-987df1309236
	google.golang.org/protobuf v1.36.7
)

require google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
