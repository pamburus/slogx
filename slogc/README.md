# slogx [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Package [slogc](https://pkg.go.dev/github.com/pamburus/slogx/slogc) provides [Logger](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Logger), which is an alias for [slogx.ContextLogger](https://pkg.go.dev/github.com/pamburus/slogx#ContextLogger) and is an alternative to [slog.Logger](https://pkg.go.dev/log/slog#Logger) that focuses on performance and usability. See the [slogx](../README.md) package description for more details about [slogx.ContextLogger](https://pkg.go.dev/github.com/pamburus/slogx#ContextLogger) and its rationale. This package provides logging primitives that are context-centric. The [Logger](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Logger) can be stored in a context using the [New](https://pkg.go.dev/github.com/pamburus/slogx/slogc#New) function, and later used implicitly by providing only the context to a set of functions such as [Log](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Log), [Info](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Info), [Debug](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Debug), and so on. It can also be retrieved from the context using the [Get](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Get) method. If the context does not contain a value stored by the [New](https://pkg.go.dev/github.com/pamburus/slogx/slogc#New) method, a [Default](https://pkg.go.dev/github.com/pamburus/slogx/slogc#Default) logger is returned, which is constructed using a handler returned by [slog.Default](https://pkg.go.dev/log/slog#Default).

## Performance
* See [benchmark results](../doc/benchmark/README.md) for details.

[doc-img]: https://pkg.go.dev/badge/github.com/pamburus/slogx/slogc
[doc]: https://pkg.go.dev/github.com/pamburus/slogx/slogc
[ci-img]: https://github.com/pamburus/slogx/actions/workflows/ci.yml/badge.svg
[ci]: https://github.com/pamburus/slogx/actions/workflows/ci.yml
[cov-img]: https://codecov.io/gh/pamburus/slogx/slogc/graph/badge.svg?token=0TF6JD4KDU
[cov]: https://codecov.io/gh/pamburus/slogx/slogc
