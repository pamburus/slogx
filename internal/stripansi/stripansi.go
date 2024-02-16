package stripansi

import (
	"regexp"
)

// Strip removes all ANSI escape sequences from the text.
func Strip(str string) string {
	return pattern.ReplaceAllString(str, "")
}

// ---

var pattern = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")