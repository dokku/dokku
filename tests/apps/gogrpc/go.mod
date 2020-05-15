module github.com/dokku/dokko/tests/apps/gorpc

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kr/pty v1.1.8 // indirect
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/mobile v0.0.0-20190814143026-e8b3e6111d02 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/tools v0.0.0-20190822000311-fc82fb2afd64 // indirect
	google.golang.org/api v0.9.0 // indirect
	google.golang.org/grpc v1.29.1
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
)

replace google.golang.org/grpc/examples/helloworld/helloworld => ./helloworld
