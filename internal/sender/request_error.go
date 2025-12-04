package sender

import "fmt"

// RequestError will be used later by other packages
type RequestError struct {
	Code    string
	Message string
}

func (e RequestError) Error() string {
	return fmt.Sprintf("Request failed: %s -> %s", e.Code, e.Message)
}
