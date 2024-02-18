// Package main is an example of how to use the slogjson package.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/slogc"
	"github.com/pamburus/slogx/slogjson"
)

func main() {
	// Create handler.
	var handler slog.Handler = slogjson.NewHandler(os.Stdout,
		slogjson.WithLevel(slog.LevelDebug-100),
		slogjson.WithSource(true),
		slogjson.WithBytesFormat(slogjson.BytesFormatString),
		slogjson.WithLevelOffset(false),
		// slogjson.WithTimeFormat("2006-01-02 15:04:05.999"),
		slogjson.WithAttrReplaceFunc(func(_ []string, attr slog.Attr) slog.Attr {
			// if len(groups) == 0 && attr.Key == slog.TimeKey {
			// 	return slog.Attr{}
			// }

			return attr
		}),
	)

	// handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// 	Level:     slog.LevelDebug,
	// 	AddSource: true,
	// })

	// Create logger and set it as default.
	slog.SetDefault(slog.New(handler))

	slog.Info("test level", slog.Any("ss", slog.LevelWarn+1), slog.Group("", slog.String("ss1", "sv1"), slog.String("ss2", "sv2")))

	slog.Info("test array", slog.Any("ss", []any{[]slog.Attr{slog.String("key1", "value1")}}))
	slog.LogAttrs(nil, slog.LevelError+10, "test level", slog.Any("ss", []byte("test-me")))

	logger := slogx.New(handler).
		With(slog.String("x", "8")).
		WithGroup("main").
		With(slog.String("y", "9")).
		WithGroup("sub").
		With(slog.Bool("z", false))
	logger.Info("start")
	logger.Error("hello there",
		slog.Any("error", errors.New("failed to figure out what to do next\ncall me later, \"friend\"")),
		slog.Int("int", 42),
		slog.Bool("bool value", true),
		slog.Any("null", nil),
		slog.Float64("float", 3.14),
		slog.Group("aaa", slog.Attr{}),
		slog.Group("oo", slog.String("oo", "false")),
	)

	// Do some logging.
	slog.Info("runtime info", slog.Int("cpu-count", runtime.NumCPU()))
	logger.Info("test string", slog.String("ss", "sv"))
	logger.Info("test level", slog.Any("ss", slog.LevelWarn+1), slog.Group("", slog.String("ss1", "sv1"), slog.String("ss2", "sv2")))
	logger.WithGroup("oo").Info("test panic", slog.Any("ss", a{}))
	slog.Info("test array", slog.Any("ss", []any{"cpu count", "abc", map[int]string{1: "one", 2: "two"}, []slog.Attr{slog.String("key1", "value1"), slog.String("key2", "value2")}}), slog.String("ss", "sv"))
	slog.Info("test empty array", slog.Any("ss", []any{}))
	slog.Info("test nil array", slog.Any("ss", []any(nil)))
	slog.Info("test array of value", slog.Any("ss", []slog.Value{slog.IntValue(1), slog.StringValue("two"), slog.AnyValue(nil), slog.BoolValue(true)}), slog.String("ss", "sv"))
	slog.Info("test bytes", slog.Any("ss", []byte("test-me")))
	slog.Warn("test=error", slogx.ErrorAttr(testError{}))
	slog.Info("", slog.Any("bytes", []byte("test-me")))
	logger.WithGroup("oo").Info("hehe", slog.Any("", []byte("test-me")))
	slog.Info("a\nb", slog.Any("bytes", []byte("test-me")))
	slog.Debug("some debug message with long field", slog.String("long-field", testError{}.Error()))
	slog.Debug(someLongText, slog.String("look at the message", "is it nice?"))

	if info, ok := debug.ReadBuildInfo(); ok {
		slog.Debug("build info", slog.String("go-version", info.GoVersion), slog.String("path", info.Path))
		for _, setting := range info.Settings {
			slog.Debug("build setting", slog.String("key", setting.Key), slog.String("value", setting.Value))
		}
		for _, dep := range info.Deps {
			slog.Debug("dependency", slog.String("path", dep.Path), slog.String("version", dep.Version))
		}
	} else {
		slog.Warn("couldn't get build info")
	}

	doSomething(context.Background())
}

func doSomething(ctx context.Context) {
	ctx = slogc.WithName(ctx, "something")

	slogc.Info(ctx, "doing something")
	slogc.Error(ctx, "something bad happened doing something", slog.Any("error", errors.New("failed to figure out what to do next")))
}

type a struct{}

func (a) String() string {
	panic(someLongText)
}

func (a) MarshalText() ([]byte, error) {
	return []byte("a"), errors.New("failed to marshal a")
}

const someLongText = "implement\nme\n" +
	`git.acronis.com/drc/go-pkg/pkg/errorx.WithStack
	/Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/errorx/error.go:23
git.acronis.com/drc/cloud-connection-service/internal/pkg/dalgorm.(*connectionDraftRef).Load
	/Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/pkg/dalgorm/connectiondraft.go:114
`

type testError struct{}

func (testError) Error() string {
	return `failed to load connection draft: record not found
record not found
git.acronis.com/drc/go-pkg/pkg/errorx.WithStack
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/errorx/error.go:23
git.acronis.com/drc/cloud-connection-service/internal/pkg/dalgorm.(*connectionDraftRef).Load
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/pkg/dalgorm/connectiondraft.go:114
git.acronis.com/drc/cloud-connection-service/internal/pkg/logic.(*connectionDraftRef).Load
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/pkg/logic/connectiondraft.go:196
git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/model.DraftRef.Load
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/model/draft.go:37
git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/handlers.connectionPost.Serve
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/handlers/connectionpost.go:46
git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/handlers.adapter.ServeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/handlers/common.go:476
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwauthz/v2.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwauthz/v2/middleware.go:45
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/cloud-connection-service/internal/pkg/tenantresolver.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/pkg/tenantresolver/middleware.go:56
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
github.com/go-chi/chi.(*ChainHandler).ServeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/chain.go:32
github.com/go-chi/chi.(*Mux).routeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/mux.go:432
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
github.com/go-chi/chi.(*Mux).ServeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/mux.go:71
github.com/go-chi/chi.(*Mux).Mount.func1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/mux.go:299
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/http/httpmetrics.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/http/httpmetrics/middleware.go:51
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwauthn.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwauthn/auth.go:92
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/clients/apigw/apigwhdr.ServerMiddleware.func1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/clients/apigw/apigwhdr/apigwhdr.go:59
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
github.com/go-chi/chi.(*ChainHandler).ServeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/chain.go:32
github.com/go-chi/chi.(*Mux).routeHTTP
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/github.com/go-chi/chi/mux.go:432
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1.RouterBuilder.Result.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/internal/unit/apiunit/api/v1/routing.go:123
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwrecovery.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwrecovery/recovery.go:69
net/http.HandlerFunc.ServeHTTP
        /usr/local/go/src/net/http/server.go:2137
git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwreqid.NewMiddleware.func1.1
        /Users/pamburus/go/src/git.acronis.com/drc/cloud-connection-service/vendor/git.acronis.com/drc/go-pkg/pkg/http/middlewares/mwreqid/requestid.go:65
`
}
