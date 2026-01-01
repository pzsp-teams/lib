package util

func MemberRole(isOwner bool) []string {
	if isOwner {
		return []string{"owner"}
	}
	return []string{}
}
