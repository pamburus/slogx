//go:build appengine || js
// +build appengine js

package ansitty

func setEnabled(uintptr, bool) bool {
	return false
}
