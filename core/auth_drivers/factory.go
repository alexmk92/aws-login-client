package auth_drivers

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexmk92/aws-login/core/types"
)

// AuthDriverName represents the authentication method to use
type AuthDriverName int

const (
	AuthDriverManual AuthDriverName = iota
	AuthDriver1Password
	AuthDriverUnknown
)

// String returns the string representation of the auth driver
func (d AuthDriverName) String() string {
	switch d {
	case AuthDriverManual:
		return "manual"
	case AuthDriver1Password:
		return "1password"
	default:
		return "unknown"
	}
}

// ParseAuthDriver parses a string to AuthDriver
func ParseAuthDriver(s string) (AuthDriverName, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "manual":
		return AuthDriverManual, nil
	case "1password":
		return AuthDriver1Password, nil
	default:
		return AuthDriverManual, fmt.Errorf("invalid auth driver '%s', valid options are: manual, 1password", s)
	}
}

// GetAuthDriverFromEnv gets the auth driver from environment variable
func GetAuthDriverFromEnv() (AuthDriverName, error) {
	driverStr := os.Getenv("AWS_LOGIN_AUTH_DRIVER")
	if driverStr == "" {
		// Default to manual if not set
		return AuthDriverManual, nil
	}

	driver, err := ParseAuthDriver(driverStr)
	if err != nil {
		return AuthDriverManual, fmt.Errorf("AWS_LOGIN_AUTH_DRIVER environment variable error: %w", err)
	}

	return driver, nil
}

// GetDriver returns the appropriate auth driver based on the driver type
func GetDriver(driverType AuthDriverName, profile string) (types.Driver, error) {
	switch driverType {
	case AuthDriverManual:
		return NewManualDriver(), nil
	case AuthDriver1Password:
		driver := NewOnePasswordDriver(profile)
		if !driver.IsInstalled() {
			return nil, fmt.Errorf("1Password CLI is not installed or not available in PATH")
		}
		return driver, nil
	default:
		return nil, fmt.Errorf("unknown auth driver: %v", driverType)
	}
}
