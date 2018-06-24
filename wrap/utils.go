package wrap

import "strings"

func RepeatWithSeparator(s string, count int, sep string) string {
	if count <= 0 {
		return ""
	}

	if count == 1 {
		return s
	}

	return s + strings.Repeat(sep+s, count-1)
}
