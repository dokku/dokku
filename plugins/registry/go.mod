module github.com/dokku/dokku/plugins/registry

go 1.16

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.5
)

replace github.com/dokku/dokku/plugins/common => ../common
