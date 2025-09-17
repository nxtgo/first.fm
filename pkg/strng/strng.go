package strng

import "unicode/utf8"

func Truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= 3 {
		return string([]rune(s)[:max])
	}
	return string([]rune(s)[:max-3]) + "..."
}
