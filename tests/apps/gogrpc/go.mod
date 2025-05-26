module github.com/dokku/dokko/tests/apps/gorpc

go 1.23.2

toolchain go1.24.1

require (
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/grpc v1.72.2
)

require (
	google.golang.org/grpc/examples v0.0.0-20240118175532-987df1309236
	google.golang.org/protobuf v1.36.6
)

require google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
