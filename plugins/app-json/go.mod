module github.com/dokku/dokku/plugins/app-json

go 1.16

require (
	github.com/dokku/dokku/plugins/apps v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
)

replace github.com/dokku/dokku/plugins/apps => ../apps

replace github.com/dokku/dokku/plugins/common => ../common
