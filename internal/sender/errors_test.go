package sender

import (
	"errors"
	"strings"
	"testing"
)

func TestRequestError_ErrorFormatsCorrectly(t *testing.T) {
	e := RequestError{
		Code:    403,
		Message: "Forbidden",
	}

	got := e.Error()
	want := "[CODE: 403]: Forbidden"

	if got != want {
		t.Fatalf("RequestError.Error() = %q, want %q", got, want)
	}
}

func TestErrData_String_Empty(t *testing.T) {
	ed := ErrData{
		ResourceRefs: map[Resource]string{},
	}

	got := ed.String()
	if got != "" {
		t.Fatalf("ErrData.String() for empty map = %q, want empty string", got)
	}
}

func TestErrData_String_SingleEntry(t *testing.T) {
	ed := ErrData{
		ResourceRefs: map[Resource]string{
			Team: "z1",
		},
	}

	got := ed.String()
	want := "TEAM(z1)"

	if got != want {
		t.Fatalf("ErrData.String() = %q, want %q", got, want)
	}
}

func TestErrData_String_MultipleEntries_NoOrderAssumed(t *testing.T) {
	ed := ErrData{
		ResourceRefs: map[Resource]string{
			Team:    "z1",
			Channel: "general",
		},
	}

	got := ed.String()

	if !strings.Contains(got, "TEAM(z1)") {
		t.Errorf("ErrData.String() = %q, expected to contain TEAM(z1)", got)
	}
	if !strings.Contains(got, "CHANNEL(general)") {
		t.Errorf("ErrData.String() = %q, expected to contain CHANNEL(general)", got)
	}

	if !strings.Contains(got, ", ") {
		t.Errorf("ErrData.String() = %q, expected to contain \", \" between entries", got)
	}
}

func TestErrAccessForbidden_ErrorIncludesCodeAndResources(t *testing.T) {
	e := ErrAccessForbidden{
		Code: 403,
		ErrData: ErrData{
			ResourceRefs: map[Resource]string{
				Team: "z1",
			},
		},
	}

	got := e.Error()
	want := "[CODE: 403]: access forbidden to one or more resources among: TEAM(z1)"

	if got != want {
		t.Fatalf("ErrAccessForbidden.Error() = %q, want %q", got, want)
	}
}

func TestErrAccessForbidden_IsMatchesSameCode(t *testing.T) {
	err := ErrAccessForbidden{
		Code: 403,
		ErrData: ErrData{
			ResourceRefs: map[Resource]string{
				Team: "z1",
			},
		},
	}

	if !errors.Is(err, ErrAccessForbidden{Code: 403}) {
		t.Fatalf("errors.Is should match ErrAccessForbidden with same code")
	}

	if errors.Is(err, ErrAccessForbidden{Code: 401}) {
		t.Fatalf("errors.Is should not match ErrAccessForbidden with different code")
	}

	if errors.Is(err, ErrResourceNotFound{Code: 403}) {
		t.Fatalf("errors.Is should not match different error type")
	}
}

func TestErrResourceNotFound_ErrorIncludesCodeAndResources(t *testing.T) {
	e := ErrResourceNotFound{
		Code: 404,
		ErrData: ErrData{
			ResourceRefs: map[Resource]string{
				Channel: "general",
			},
		},
	}

	got := e.Error()
	want := "[CODE: 404]: one or more resources not found among: CHANNEL(general)"

	if got != want {
		t.Fatalf("ErrResourceNotFound.Error() = %q, want %q", got, want)
	}
}

func TestErrResourceNotFound_IsMatchesSameCode(t *testing.T) {
	err := ErrResourceNotFound{
		Code: 404,
		ErrData: ErrData{
			ResourceRefs: map[Resource]string{
				Channel: "general",
			},
		},
	}

	if !errors.Is(err, ErrResourceNotFound{Code: 404}) {
		t.Fatalf("errors.Is should match ErrResourceNotFound with same code")
	}

	if errors.Is(err, ErrResourceNotFound{Code: 400}) {
		t.Fatalf("errors.Is should not match ErrResourceNotFound with different code")
	}

	if errors.Is(err, ErrAccessForbidden{Code: 404}) {
		t.Fatalf("errors.Is should not match different error type")
	}
}
