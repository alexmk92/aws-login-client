package auth_drivers

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/alexmk92/aws-login/core"
	"github.com/alexmk92/aws-login/core/types"
)

// OnePasswordDriver implements 1Password MFA token retrieval
type OnePasswordDriver struct {
	vaultKey string
	profile  string
}

// This is a type assertion to the compiler to ensure that OnePasswordDriver implements the Driver interface
// if the constraints aren't met, the compiler will throw an error
//
// We do this because we have a @factory.go file that returns a types.Driver interface and we need to
// ensure that if someone bypases the factory and tries to create a OnePasswordDriver directly,
// the compiler will throw an error
var _ types.Driver = (*OnePasswordDriver)(nil)

// NewOnePasswordDriver creates a new 1Password driver
func NewOnePasswordDriver(profile string) *OnePasswordDriver {
	credentialReader := core.GetCredentialReader()
	credential, exists := credentialReader.GetCredential(profile)

	vaultKey := ""
	if exists {
		vaultKey = credential.VaultKey
	}

	return &OnePasswordDriver{vaultKey: vaultKey, profile: profile}
}

// GetToken retrieves MFA token from 1Password
func (d *OnePasswordDriver) GetToken() (string, error) {
	cmd := exec.Command("op", "item", "get", d.vaultKey, "--otp")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to retrieve MFA code from 1Password: %w", err)
	}

	mfaCode := strings.TrimSpace(string(output))
	if mfaCode == "" {
		return "", fmt.Errorf("empty MFA code from 1Password")
	}

	return mfaCode, nil
}

// Name returns the name of the driver
func (d *OnePasswordDriver) Name() string {
	return "1password"
}

func (d *OnePasswordDriver) YieldsMFACode() bool {
	return true
}

func (d *OnePasswordDriver) GetMFACode() (string, error) {
	cmd := exec.Command("op", "item", "get", d.vaultKey, "--otp")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to retrieve MFA code from 1Password with vault key %s: %w", d.vaultKey, err)
	}

	mfaCode := strings.TrimSpace(string(output))
	if mfaCode == "" {
		return "", fmt.Errorf("empty MFA code from 1Password")
	}

	return mfaCode, nil
}

func (d OnePasswordDriver) IsInstalled() bool {
	cmd := exec.Command("op", "--version")
	err := cmd.Run()
	return err == nil
}
