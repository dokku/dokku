module github.com/dokku/dokku/plugins/ps

go 1.19

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/config v0.0.0-00010101000000-000000000000
	github.com/dokku/dokku/plugins/docker-options v0.0.0-00010101000000-000000000000
	github.com/gofrs/flock v0.8.0
	github.com/otiai10/copy v1.9.0
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/joho/godotenv v1.2.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common

replace github.com/dokku/dokku/plugins/config => ../config

replace github.com/dokku/dokku/plugins/docker-options => ../docker-options
