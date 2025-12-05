package mapper

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
