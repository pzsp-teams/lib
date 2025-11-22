package sender

import "fmt"

type RequestError struct {
	Code    string
	Message string
}

func (e RequestError) Error() string {
	return fmt.Sprintf("Request failed: %s -> %s", e.Code, e.Message)
}
