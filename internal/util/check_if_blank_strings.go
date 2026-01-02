package util

import (
	"strings"
)

func CheckIfAnyStringIsBlank(strs ...string) bool {
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str == "" {
			return true
		}
	}
	return false
}