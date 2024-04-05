module github.com/dokku/dokku/plugins/app-json

go 1.21

require (
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.7.0
	k8s.io/utils v0.0.0-20240102154912-e7106e64919e
)

require (
	github.com/alexellis/go-execute/v2 v2.2.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/melbahja/goph v1.4.0 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.5 // indirect
	github.com/ryanuber/columnize v2.1.2+incompatible // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
