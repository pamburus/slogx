package slogjson_test

import (
	"log/slog"
	"testing"
	"time"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx/slogjson"
)

func TestDuration(tt *testing.T) {
	t := New(tt)

	t.Run("AsText", func(t Test) {
		buf, v := slogjson.DurationAsText()(nil, 42*time.Second)
		t.Expect(buf).To(BeNil())
		t.Expect(v.Kind()).To(Equal(slog.KindString))
		t.Expect(v.String()).To(Equal("42s"))
	})

	t.Run("AsSeconds", func(t Test) {
		buf, v := slogjson.DurationAsSeconds()(nil, time.Minute+time.Second)
		t.Expect(buf).To(BeNil())
		t.Expect(v.Kind()).To(Equal(slog.KindFloat64))
		t.Expect(v.Float64()).To(Equal(61.0))
	})

	t.Run("AsHMS", func(t Test) {
		buf, v := slogjson.DurationAsHMS()(nil, time.Hour+16*time.Minute+40*time.Millisecond)
		t.Expect(buf).To(Equal([]byte("01:16:00.04")))
		t.Expect(v.Equal(slog.Value{})).To(BeTrue())

		buf, v = slogjson.DurationAsHMS(slogjson.WithDurationPrecision(3))(nil, time.Hour+16*time.Minute+40*time.Millisecond)
		t.Expect(buf).To(Equal([]byte("01:16:00.040")))
		t.Expect(v.Equal(slog.Value{})).To(BeTrue())
	})
}
