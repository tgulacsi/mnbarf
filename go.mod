module github.com/tgulacsi/mnbarf

require (
	github.com/cockroachdb/apd v1.1.0
	github.com/hooklift/gowsdl/generator v0.0.0-596224a774ef2fabb724cd8a955e53006beac705
	github.com/pkg/errors v0.8.0
)

replace github.com/hooklift/gowsdl/generator v0.0.0-596224a774ef2fabb724cd8a955e53006beac705 => ./vendor/github.com/hooklift/gowsdl/generator
