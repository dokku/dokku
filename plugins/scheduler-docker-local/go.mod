module github.com/dokku/dokku/plugins/scheduler-docker-local

go 1.26.2

require (
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/cron v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/storage v0.0.0-00010101000000-000000000000
	github.com/multiformats/go-base36 v0.2.0
	github.com/spf13/pflag v1.0.10
	golang.org/x/sync v0.20.0
)

require (
	github.com/alexellis/go-execute/v2 v2.2.1 // indirect
	github.com/dokku/dokku/plugins/app-json v0.0.0-00010101000000-000000000000 // indirect
	github.com/dokku/dokku/plugins/docker-options v0.0.0-00010101000000-000000000000 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/melbahja/goph v1.5.0 // indirect
	github.com/otiai10/copy v1.14.1 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/ryanuber/columnize v2.1.2+incompatible // indirect
	github.com/tailscale/hujson v0.0.0-20250605163823-992244df8c5a // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	k8s.io/utils v0.0.0-20240102154912-e7106e64919e // indirect
	mvdan.cc/sh/v3 v3.13.1 // indirect
)

replace github.com/dokku/dokku/plugins/app-json => ../app-json

replace github.com/dokku/dokku/plugins/common => ../common

replace github.com/dokku/dokku/plugins/cron => ../cron

replace github.com/dokku/dokku/plugins/storage => ../storage

replace github.com/dokku/dokku/plugins/docker-options => ../docker-options
