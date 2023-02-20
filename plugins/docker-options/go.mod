module github.com/dokku/dokku/plugins/docker-options

go 1.19

require github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000

require (
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/otiai10/copy v1.9.0 // indirect
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
