package sender

import "errors"

type StatusCoder interface {
	error
	StatusCode() int
}

func StatusCode(err error) (int, bool) {
	var sc StatusCoder
	if errors.As(err, &sc) {
		return sc.StatusCode(), true
	}
	return 0, false
}

func (e RequestError) StatusCode() int {
	return e.Code
}

func (e ErrAccessForbidden) StatusCode() int {
	return e.Code
}

func (e ErrResourceNotFound) StatusCode() int {
	return e.Code
}
