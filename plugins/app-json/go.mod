module github.com/dokku/dokku/plugins/app-json

go 1.15

require (
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27 // indirect
	github.com/dokku/dokku/plugins/common v0.0.0-00010101000000-000000000000
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/ryanuber/columnize v1.1.2-0.20190319233515-9e6335e58db3 // indirect
)

replace github.com/dokku/dokku/plugins/common => ../common
