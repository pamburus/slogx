module github.com/pamburus/slogx/slogtext/example

go 1.22.0

replace github.com/pamburus/slogx => ../..

replace github.com/pamburus/slogx/ansitty => ../../ansitty

require (
	github.com/pamburus/slogx v0.0.0-00010101000000-000000000000
	github.com/pamburus/slogx/ansitty v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.17.0 // indirect
