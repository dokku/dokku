module github.com/dokku/dokku/plugins/scheduler

go 1.16

require (
	github.com/dokku/dokku/plugins/apps v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/config v0.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.5
)

replace github.com/dokku/dokku/plugins/apps => ../apps

replace github.com/dokku/dokku/plugins/common => ../common

replace github.com/dokku/dokku/plugins/config => ../config
