package  main

import (
	"fmt"
)

func safeString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func safeInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func safeScopes(m map[string]interface{}) ([]string, error) {
	raw, ok := m["scopes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid scopes format")
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid scope value")
		}
		out = append(out, s)
	}
	return out, nil
}

func validateAuthMethod(method string) (string, error) {
	if method == "DEVICE_CODE" || method == "INTERACTIVE" {
		return method, nil
	}
	return "", fmt.Errorf("invalid auth method: %s", method)
}