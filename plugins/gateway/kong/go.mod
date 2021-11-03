module github.com/apiclarity/apiclarity/plugins/gateway/kong

go 1.16

require github.com/Kong/go-pdk v0.7.1

require (
	github.com/apiclarity/apiclarity/plugins/api v0.0.0
	github.com/go-openapi/runtime v0.21.0
	github.com/go-openapi/strfmt v0.21.0
)

replace github.com/apiclarity/apiclarity/plugins/api v0.0.0 => ./../../api
