module github.com/dokku/dokku/plugins/builder

go 1.17

require (
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/otiai10/copy v1.6.0
	github.com/spf13/pflag v1.0.5
)

replace github.com/dokku/dokku/plugins/common => ../common
