module github.com/dokku/dokku/plugins/logs

go 1.17

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/docker-options v0.0.0-00010101000000-000000000000
	github.com/joncalhoun/qson v0.0.0-20200422171543-84433dcd3da0
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common

replace github.com/dokku/dokku/plugins/docker-options => ../docker-options
