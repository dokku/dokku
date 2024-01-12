module github.com/dokku/dokku/plugins/apps

go 1.21

require (
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	github.com/ryanuber/columnize v2.1.2+incompatible // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
