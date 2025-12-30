package sender

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/resources"
)

var errTemplate string = "[CODE: %d]: %s"

type RequestError struct {
	Code    int
	Message string
}

func (e RequestError) Error() string {
	return fmt.Sprintf(errTemplate, e.Code, e.Message)
}

type ErrData struct {
	ResourceRefs map[resources.Resource]string
}

func (ed *ErrData) String() string {
	formattedRefs := make([]string, 0, len(ed.ResourceRefs))
	for t, ref := range ed.ResourceRefs {
		formattedRefs = append(formattedRefs, fmt.Sprintf("%s(%s)", t, ref))
	}
	return strings.Join(formattedRefs, ", ")
}

type ErrAccessForbidden struct {
	Code int
	ErrData
}

func (e ErrAccessForbidden) Error() string {
	return fmt.Sprintf(
		errTemplate,
		e.Code,
		fmt.Sprintf("access forbidden to one or more resources among: %s", e.String()),
	)
}

func (e ErrAccessForbidden) Is(target error) bool {
	t, ok := target.(ErrAccessForbidden)
	return ok && e.Code == t.Code
}

type ErrResourceNotFound struct {
	Code int
	ErrData
}

func (e ErrResourceNotFound) Error() string {
	return fmt.Sprintf(
		errTemplate,
		e.Code,
		fmt.Sprintf("one or more resources not found among: %s", e.String()),
	)
}

func (e ErrResourceNotFound) Is(target error) bool {
	t, ok := target.(ErrResourceNotFound)
	return ok && e.Code == t.Code
}
