package resolver

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/resources"
)

type resourcesNotAvailableError struct {
	resourceType resources.Resource
}

func (e *resourcesNotAvailableError) Error() string {
	return fmt.Sprintf("cannot resolve %s: resources not available", e.resourceType)
}

type resourceNotFoundError struct {
	resourceType resources.Resource
	ref          string
}

func (e *resourceNotFoundError) Error() string {
	return fmt.Sprintf("%s referenced by %q not found", e.resourceType, e.ref)
}

type resourceEmptyIDError struct {
	resourceType resources.Resource
	ref          string
}

func (e *resourceEmptyIDError) Error() string {
	return fmt.Sprintf("%s referenced by %q has empty ID", e.resourceType, e.ref)
}

type resourceAmbiguousError struct {
	resourceType resources.Resource
	ref          string
	options      []string
}

func (e *resourceAmbiguousError) Error() string {
	return fmt.Sprintf("multiple %ss referenced by %q found: %s\n. \nPlease use one of the IDs instead",
		e.resourceType, e.ref, strings.Join(e.options, ";\n"))
}
