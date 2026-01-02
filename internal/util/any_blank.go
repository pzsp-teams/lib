package util

import (
	"strings"
)

func AnyBlank(strs ...string) bool {
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str == "" {
			return true
		}
	}
	return false
}