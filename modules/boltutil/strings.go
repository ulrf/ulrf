package boltutil

import (
	"strings"
)

func cleanString(in string) (res string) {
	clean := func(r rune) bool {
		if ('а' <= r &&
			'я' >= r) || r == 'ё' {
			return false
		}
		return true
	}

	in = strings.ToLower(in)
	arr := strings.FieldsFunc(in, clean)
	for _, word := range arr {
		if len([]rune(word)) > 2 {
			res += word
		}
	}
	return strings.TrimSpace(res)
}
