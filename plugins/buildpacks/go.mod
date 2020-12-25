module github.com/dokku/dokku/plugins/buildpacks

go 1.15

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
