module github.com/pamburus/slogx/cmd/slogxfmt

go 1.22.0

replace github.com/pamburus/slogx => ../..

require (
	github.com/alexflint/go-arg v1.4.3
	github.com/pamburus/ansitty v0.1.0
	github.com/pamburus/slogx v0.0.0-00010101000000-000000000000
	github.com/valyala/fastjson v1.6.4
)

require (
	github.com/alexflint/go-scalar v1.1.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)
