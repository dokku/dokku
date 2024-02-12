module github.com/dokku/dokku/plugins/cron

go 1.21

require (
	github.com/dokku/dokku/plugins/app-json v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/multiformats/go-base36 v0.2.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/ryanuber/columnize v2.1.2+incompatible
	github.com/spf13/pflag v1.0.5
	mvdan.cc/sh/v3 v3.8.0
)

require (
	github.com/alexellis/go-execute/v2 v2.2.1 // indirect
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/melbahja/goph v1.4.0 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.5 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace github.com/dokku/dokku/plugins/app-json => ../app-json

replace github.com/dokku/dokku/plugins/common => ../common
