package util

import (
	"regexp"
	"strings"
)

func IsLikelyThreadConversationID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@thread.")
}

func IsLikelyChatID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@unq.")
}

func IsLikelyGUID(s string) bool {
	var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return guidRegex.MatchString(s)
}
