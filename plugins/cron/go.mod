module github.com/dokku/dokku/plugins/cron

go 1.19

require (
	github.com/dokku/dokku/plugins/app-json v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/multiformats/go-base36 v0.1.1-0.20220823151017-f5af2eed4d9c
	github.com/robfig/cron/v3 v3.0.1
	github.com/ryanuber/columnize v2.1.2+incompatible
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/otiai10/copy v1.9.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)

replace github.com/dokku/dokku/plugins/app-json => ../app-json

replace github.com/dokku/dokku/plugins/common => ../common
