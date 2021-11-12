module github.com/dokku/dokku/plugins/logs

go 1.16

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/apps v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/docker-options v0.0.0-00010101000000-000000000000
	github.com/joncalhoun/qson v0.0.0-20200422171543-84433dcd3da0
	github.com/spf13/pflag v1.0.5
)

replace github.com/dokku/dokku/plugins/apps => ../apps

replace github.com/dokku/dokku/plugins/common => ../common

replace github.com/dokku/dokku/plugins/docker-options => ../docker-options
