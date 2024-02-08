package slogc_test

import (
	"context"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx/slogc"

	"testing"
)

func TestNewGet(tt *testing.T) {
	t := New(tt)

	t.Run("New", func(t Test) {
		ctx := context.Background()
		ctx = slogc.New(ctx, slogc.Default())
	})
}
