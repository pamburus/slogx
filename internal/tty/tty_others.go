//go:build appengine || js
// +build appengine js

package tty

import "errors"

func enableSeqTTY(fd uintptr, flag bool) error {
	return errors.New("default not a terminal")
}
