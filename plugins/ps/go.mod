module github.com/dokku/dokku/plugins/ps

go 1.14

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
  github.com/dokku/dokku/plugins/config v0.0.0-00010101000000-000000000000
  github.com/dokku/dokku/plugins/docker-options v0.0.0-00010101000000-000000000000
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3
)

replace github.com/dokku/dokku/plugins/common => ../common
replace github.com/dokku/dokku/plugins/config => ../config
replace github.com/dokku/dokku/plugins/docker-options => ../docker-options
