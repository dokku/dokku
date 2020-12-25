module github.com/dokku/dokku/plugins/repo

go 1.15

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
  github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
  github.com/dokku/dokku/plugins/config v0.0.0-00010101000000-000000000000
  github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
replace github.com/dokku/dokku/plugins/config => ../config
