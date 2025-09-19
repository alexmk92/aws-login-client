package auth_drivers

import (
	"fmt"

	"github.com/alexmk92/aws-login/core/types"
)

// ManualDriver implements manual MFA token input
type ManualDriver struct{}

// This is a type assertion to the compiler to ensure that ManualDriver implements the Driver interface
// if the constraints aren't met, the compiler will throw an error
//
// We do this because we have a @factory.go file that returns a types.Driver interface and we need to
// ensure that if someone bypases the factory and tries to create a ManualDriver directly,
// the compiler will throw an error
var _ types.Driver = (*ManualDriver)(nil)

// NewManualDriver creates a new manual driver
func NewManualDriver() *ManualDriver {
	return &ManualDriver{}
}

// GetToken prompts the user for manual MFA token input
func (d *ManualDriver) GetToken() (string, error) {
	// This will be called from the UI layer, so we'll use a simple approach
	// The actual UI interaction will be handled in the UI layer
	return "", fmt.Errorf("manual token input should be handled by UI layer")
}

// Name returns the name of the driver
func (d *ManualDriver) Name() string {
	return "manual"
}

// ValidateMFACode checks if the MFA code is 6 digits
func (d *ManualDriver) ValidateMFACode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func (d *ManualDriver) YieldsMFACode() bool {
	return false
}

func (d *ManualDriver) GetMFACode() (string, error) {
	return "", fmt.Errorf("manual driver cannot yield MFA code")
}

func (d *ManualDriver) IsInstalled() bool {
	return true
}
