module github.com/dokku/dokko/tests/apps/gorpc

go 1.22
toolchain go1.23.0

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/grpc v1.68.0
)

require (
	google.golang.org/grpc/examples v0.0.0-20240118175532-987df1309236
	google.golang.org/protobuf v1.35.1
)

require google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
