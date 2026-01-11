package setup

import "github.com/pzsp-teams/lib/internal/pepper"

// PepperExists checks whether a pepper is set in the system keyring.
func PepperExists() bool {
	exists, err := pepper.PepperExists()
	if err != nil {
		return false
	}
	return exists
}

// SetPepper sets the given pepper in the system keyring.
func SetPepper(pep string) error {
	return pepper.SetPepper(pep)
}
