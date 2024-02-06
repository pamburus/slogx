# slogx [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Module [slogx](https://pkg.go.dev/github.com/pamburus/slogx) provides extensions and helpers for the [log/slog](https://pkg.go.dev/log/slog) package.

## Packages
* [slogx](./README.md)
* [slogc](slogc/README.md)


### Package slogx

Package [slogx](https://pkg.go.dev/github.com/pamburus/slogx) provides [Logger](https://pkg.go.dev/github.com/pamburus/slogx#Logger) as an alternative to [slog.Logger](https://pkg.go.dev/log/slog#Logger), which focuses on performance and makes several design changes to improve it:
* It does not provide convenience methods for attributes that may affect performance. All methods accept attributes only as [slog.Attr](https://pkg.go.dev/log/slog#Attr).
* It provides an [option](https://pkg.go.dev/github.com/pamburus/slogx#Logger.WithSource) to disable the inclusion of [source](https://pkg.go.dev/log/slog#Source) information in the log [Record](https://pkg.go.dev/log/slog#Record). This can improve performance by up to 100% in cases where the source information is not included by the [Handler](https://pkg.go.dev/log/slog#Handler) anyway.
* Its [With](https://pkg.go.dev/github.com/pamburus/slogx#Logger.With) method does not immediately call the handler's [WithAttrs](https://pkg.go.dev/log/slog#Handler.WithAttrs) method, instead it buffers up to 4 attributes which are then added to each log [Record](https://pkg.go.dev/log/slog#Record). This improves performance when you need to define a temporary set of attributes in a function and log a few messages with those attributes a few times. It also greatly improves performance when the logger is disabled. This is because calling [WithAttrs](https://pkg.go.dev/log/slog#Handler.WithAttrs) is quite an expensive operation, especially if the [Handler](https://pkg.go.dev/log/slog#Handler) is wrapped multiple times. That is, each layer will call the underlying handler's [WithAttrs](https://pkg.go.dev/log/slog#Handler.WithAttrs) method and cause a lot of allocations. But what if the message is discarded because the logger is disabled? Yes, it will be a waste of CPU time. So for temporary [With](https://pkg.go.dev/github.com/pamburus/slogx#Logger.With) attribute sets, it is usually more efficient to keep them on the [Logger](https://pkg.go.dev/github.com/pamburus/slogx#Logger). If any of the [WithGroup](https://pkg.go.dev/github.com/pamburus/slogx#Logger.WithGroup), [Handler](https://pkg.go.dev/github.com/pamburus/slogx#Logger.Handler), or [LongTerm](https://pkg.go.dev/github.com/pamburus/slogx#Logger.LongTerm) methods are called later, the temporary attributes will be flushed using the [WithAttrs](https://pkg.go.dev/log/slog#Handler.WithAttrs) method.
* It provides the [WithLongTerm](https://pkg.go.dev/github.com/pamburus/slogx#Logger.WithLongTerm) method, which acts as a sequence of [With](https://pkg.go.dev/github.com/pamburus/slogx#Logger.With) and [LongTerm](https://pkg.go.dev/github.com/pamburus/slogx#Logger.LongTerm) method calls and is needed for cases where the resulting logger is intended to be reused multiple times and may reside in a rather long-lived context.

## Performance
* See [benchmark results](doc/benchmark/README.md) for details.

[doc-img]: https://pkg.go.dev/badge/github.com/pamburus/slogx
[doc]: https://pkg.go.dev/github.com/pamburus/slogx
[ci-img]: https://github.com/pamburus/slogx/actions/workflows/ci.yml/badge.svg
[ci]: https://github.com/pamburus/slogx/actions/workflows/ci.yml
[cov-img]: https://codecov.io/gh/pamburus/slogx/graph/badge.svg?token=0TF6JD4KDU
[cov]: https://codecov.io/gh/pamburus/slogx
