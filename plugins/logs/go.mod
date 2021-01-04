module github.com/dokku/dokku/plugins/logs

go 1.15

require (
	github.com/RaveNoX/go-jsonmerge v1.0.0
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/joncalhoun/qson v0.0.0-20200422171543-84433dcd3da0
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xo/dburl v0.1.0
)

replace github.com/dokku/dokku/plugins/common => ../common
