package slogjson_test

import (
	"log/slog"
	"testing"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx/slogjson"
)

func TestSource(tt *testing.T) {
	t := New(tt)

	t.Run("SourceLongObject", func(t Test) {
		buf, v := slogjson.SourceLongObject()(nil, slog.Source{
			Line:     42,
			File:     "/some/test/file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(BeNil())
		t.Expect(v.Kind()).To(Equal(slog.KindGroup))
		t.Expect(v.Group()).To(HaveLen(3))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("file")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.StringValue("/some/test/file.go"))),
		)))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("line")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.IntValue(42))),
		)))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("function")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.StringValue("someFunction"))),
		)))
	})

	t.Run("SourceShortObject", func(t Test) {
		buf, v := slogjson.SourceShortObject()(nil, slog.Source{
			Line:     42,
			File:     "/some/test/file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(BeNil())
		t.Expect(v.Kind()).To(Equal(slog.KindGroup))
		t.Expect(v.Group()).To(HaveLen(3))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("file")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.StringValue("test/file.go"))),
		)))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("line")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.IntValue(42))),
		)))
		t.Expect(v.Group()).To(Contain(Struct(
			Field("Key", Equal("function")),
			Field("Value", EqualUsing(slog.Value.Equal, slog.StringValue("someFunction"))),
		)))
	})

	t.Run("SourceLongString", func(t Test) {
		buf, v := slogjson.SourceLongString()(nil, slog.Source{
			Line:     42,
			File:     "/some/test/file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(Equal([]byte("/some/test/file.go:42")))
		t.Expect(v).To(BeZero())
	})

	t.Run("SourceShortString", func(t Test) {
		buf, v := slogjson.SourceShortString()(nil, slog.Source{
			Line:     42,
			File:     "/some/test/file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(Equal([]byte("test/file.go:42")))
		t.Expect(v).To(BeZero())
	})

	t.Run("WithoutDirs", func(t Test) {
		buf, v := slogjson.SourceShortString()(nil, slog.Source{
			Line:     42,
			File:     "file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(Equal([]byte("file.go:42")))
		t.Expect(v).To(BeZero())
	})

	t.Run("WithOneDir", func(t Test) {
		buf, v := slogjson.SourceShortString()(nil, slog.Source{
			Line:     42,
			File:     "test/file.go",
			Function: "someFunction",
		})

		t.Expect(buf).To(Equal([]byte("test/file.go:42")))
		t.Expect(v).To(BeZero())
	})
}
