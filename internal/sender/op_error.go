package sender

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/resources"
)

type Param struct {
	Key   resources.Key
	Value []string
}

type OpError struct {
	Operation string
	Params    []Param
	Err       error
}

func (e *OpError) Unwrap() error {
	return e.Err
}

func (e *OpError) Error() string {
	if len(e.Params) == 0 {
		return fmt.Sprintf("Error in %s: %v", e.Operation, e.Err)
	}
	parts := make([]string, 0, len(e.Params))
	for _, p := range e.Params {
		parts = append(parts, fmt.Sprintf("%s=%s", p.Key, p.Value))
	}
	return fmt.Sprintf("Error in %s [%s]: %v", e.Operation, strings.Join(parts, ", "), e.Err)
}

func Wrap(op string, err error, params ...Param) error {
	if err == nil {
		return nil
	}
	return &OpError{
		Operation: op,
		Params:    params,
		Err:       err,
	}
}