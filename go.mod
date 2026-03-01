module github.com/tgulacsi/mnbarf

go 1.25.0

require (
	github.com/UNO-SOFT/zlog v0.8.6
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/peterbourgon/ff/v3 v3.4.0
	github.com/rogpeppe/retry v0.1.0
	github.com/valyala/quicktemplate v1.8.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/hooklift/gowsdl v0.5.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/term v0.40.0 // indirect
)

tool (
	github.com/hooklift/gowsdl/cmd/gowsdl
	github.com/valyala/quicktemplate/qtc
)
