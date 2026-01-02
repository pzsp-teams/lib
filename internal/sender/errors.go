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
	ResourceRefs map[resources.Resource][]string
}

func (ed *ErrData) String() string {
	formatted := make([]string, 0, len(ed.ResourceRefs))
	for t, refs := range ed.ResourceRefs {
		formatted = append(formatted, fmt.Sprintf("%s(%s)", t, strings.Join(refs, ",")))
	}
	return strings.Join(formatted, ", ")
}


type ErrAccessForbidden struct {
	Code            int
	OriginalMessage string
	ErrData
}

func (e ErrAccessForbidden) Error() string {
	return fmt.Sprintf(
		errTemplate,
		e.Code,
		fmt.Sprintf("access forbidden to one or more resources among: %s (%s)", e.String(), e.OriginalMessage),
	)
}

func (e ErrAccessForbidden) Is(target error) bool {
	t, ok := target.(ErrAccessForbidden)
	return ok && e.Code == t.Code
}

type ErrResourceNotFound struct {
	Code            int
	OriginalMessage string
	ErrData
}

func (e ErrResourceNotFound) Error() string {
	return fmt.Sprintf(
		errTemplate,
		e.Code,
		fmt.Sprintf("one or more resources not found among: %s (%s)", e.String(), e.OriginalMessage),
	)
}

func (e ErrResourceNotFound) Is(target error) bool {
	t, ok := target.(ErrResourceNotFound)
	return ok && e.Code == t.Code
}
