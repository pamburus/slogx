package slogjson

import (
	"log/slog"
	"testing"

	. "github.com/pamburus/go-tst/tst"
)

func TestOptions(tt *testing.T) {
	t := New(tt)

	options := func(options ...Option) options {
		return defaultOptions().with(options)
	}

	t.Run("nil", func(t Test) {
		options(nil)
	})

	t.Run("WithLevel", func(t Test) {
		t.Run("default", func(t Test) {
			leveler := options().leveler
			t.Expect(leveler).ToNot(BeNil())
			t.Expect(leveler).To(Equal(slog.LevelInfo))
		})
		t.Run("nil", func(t Test) {
			leveler := options(WithLevel(nil)).leveler
			t.Expect(leveler).ToNot(BeNil())
			t.Expect(leveler).To(Equal(slog.LevelInfo))
		})
		t.Run("non-nil", func(t Test) {
			leveler := options(WithLevel(slog.LevelDebug)).leveler
			t.Expect(leveler).To(Equal(slog.LevelDebug))
		})
	})

	t.Run("WithAttrReplaceFunc", func(t Test) {
		t.Run("default", func(t Test) {
			replaceAttr := options().replaceAttr
			t.Expect(replaceAttr).To(BeNil())
		})
		t.Run("non-nil", func(t Test) {
			inputGroups := []string{"a", "b", "c"}
			inputAttr := slog.Int("ki", 42)
			goldenOutputAttr := slog.String("ko", "42")
			f := func(groups []string, attr slog.Attr) slog.Attr {
				t.Expect(groups).ToEqual(inputGroups)
				t.Expect(attr).ToEqual(inputAttr)

				return goldenOutputAttr
			}
			replaceAttr := options(WithAttrReplaceFunc(f)).replaceAttr
			t.Expect(replaceAttr).ToNot(BeNil())
			outputAttr := replaceAttr(inputGroups, inputAttr)
			t.Expect(outputAttr.Key).ToEqual(goldenOutputAttr.Key)
			t.Expect(outputAttr.Value.Kind()).ToEqual(goldenOutputAttr.Value.Kind())
			t.Expect(outputAttr.Value.String()).ToEqual(goldenOutputAttr.Value.String())
		})
	})

	t.Run("WithSource", func(t Test) {
		t.Run("default", func(t Test) {
			includeSource := options().includeSource
			t.Expect(includeSource).To(BeFalse())
		})
		t.Run("true", func(t Test) {
			t.Expect(options(WithSource(true)).includeSource).To(BeTrue())
		})
		t.Run("false", func(t Test) {
			t.Expect(options(WithSource(false)).includeSource).To(BeFalse())
		})
	})

	t.Run("WithSourceEncodeFunc", func(t Test) {
		t.Run("default", func(t Test) {
			encodeSource := options().encodeSource
			t.Expect(encodeSource).ToNot(BeNil())
			buf, source := encodeSource(nil, slog.Source{
				File:     "/some/long/file",
				Line:     42,
				Function: "function",
			})
			t.Expect(buf).To(BeNil())
			t.Expect(source.Kind()).To(Equal(slog.KindGroup))
			t.Expect(source.String()).To(Equal("[file=/some/long/file line=42 function=function]"))
		})
		t.Run("non-nil", func(t Test) {
			inputBuf := []byte("input")
			inputSource := slog.Source{
				File:     "/some/long/file",
				Line:     42,
				Function: "function",
			}
			goldenOutputBuf := []byte("output")
			goldenOutputValue := slog.StringValue("long/file:42")
			f := func(buf []byte, source slog.Source) ([]byte, slog.Value) {
				t.Expect(buf).ToEqual(inputBuf)
				t.Expect(source).ToEqual(inputSource)

				return goldenOutputBuf, goldenOutputValue
			}
			encodeSource := options(WithSourceEncodeFunc(f)).encodeSource
			t.Expect(encodeSource).ToNot(BeNil())
			outputBuf, outputValue := encodeSource(inputBuf, inputSource)
			t.Expect(outputBuf).ToEqual(goldenOutputBuf)
			t.Expect(outputValue.Kind()).ToEqual(goldenOutputValue.Kind())
			t.Expect(outputValue.String()).ToEqual(goldenOutputValue.String())
		})
	})
}
