package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexmk92/aws-login/core/types"
)

// CredentialReader handles reading and parsing AWS credentials file
type CredentialReader struct {
	credentials      map[string]types.StaticCredential
	roleArnToProfile map[string]string // Maps role ARN to profile name for quick lookup
}

// Make this a doOnce singleton
var credentialReaderInstance *CredentialReader
var credentialReaderOnce sync.Once

func NewCredentialReader() *CredentialReader {
	credentialReaderOnce.Do(func() {
		credentialReaderInstance = &CredentialReader{
			credentials:      make(map[string]types.StaticCredential),
			roleArnToProfile: make(map[string]string),
		}
	})
	return credentialReaderInstance
}

func GetCredentialReader() *CredentialReader {
	if credentialReaderInstance == nil {
		NewCredentialReader()
	}

	return credentialReaderInstance
}

// LoadCredentialsFile loads and parses the AWS credentials file
func (cr *CredentialReader) LoadCredentialsFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	credentialsPath := filepath.Join(homeDir, ".aws", "credentials")
	file, err := os.Open(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to open credentials file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentProfile string
	var currentCredential types.StaticCredential

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for profile header [profile_name]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous profile if it exists
			if currentProfile != "" {
				cr.credentials[currentProfile] = currentCredential
				// Add to role ARN lookup map if this profile has an assumable role
				if currentCredential.AssumableRoleID != "" {
					cr.roleArnToProfile[currentCredential.AssumableRoleID] = currentProfile
				}
			}

			// Start new profile
			currentProfile = strings.Trim(line, "[]")
			currentCredential = types.StaticCredential{
				ProfileName: currentProfile,
			}
			continue
		}

		// Parse key-value pairs
		if currentProfile != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Skip empty values
				if value == "" {
					continue
				}

				switch key {
				case "aws_access_key_id":
					currentCredential.AccessKey = value
				case "aws_secret_access_key":
					currentCredential.AccessSecret = value
				case "account_id", "aws_account_id":
					currentCredential.AccountID = value
				case "mfa_serial":
					currentCredential.MfaSerial = value
				case "assumable_role_id":
					currentCredential.AssumableRoleID = value
				case "vault_key":
					currentCredential.VaultKey = value
				}
			}
		}
	}

	// Save the last profile
	if currentProfile != "" {
		cr.credentials[currentProfile] = currentCredential
		// Add to role ARN lookup map if this profile has an assumable role
		if currentCredential.AssumableRoleID != "" {
			cr.roleArnToProfile[currentCredential.AssumableRoleID] = currentProfile
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading credentials file: %w", err)
	}

	return nil
}

// Returns a list of all profile names that we can attempt to assume a role
// for.  If we only define the vault key or role arn, then we don't want
// to include is as an authable entity.  It could however still be consumed
// by another profile (such as prd acting as int via an assumable role)
func (cr *CredentialReader) GetValidProfiles() []string {
	profiles := make([]string, 0, len(cr.credentials))

	for profile, credential := range cr.credentials {
		if credential.AccessKey != "" && credential.AccessSecret != "" && credential.MfaSerial != "" {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

// GetCredential returns the credential for a specific profile
func (cr *CredentialReader) GetCredential(profile string) (types.StaticCredential, bool) {
	credential, exists := cr.credentials[profile]
	return credential, exists
}

// GetAssumableRoles returns the list of roles that can be assumed for a profile
// This now returns all profiles that have an assumable_role_id (except the current profile)
func (cr *CredentialReader) GetAssumableRoles(profile string) []string {
	var assumableRoles []string

	for profileName, credential := range cr.credentials {
		// Skip the current profile - it can't assume itself
		if profileName == profile {
			continue
		}

		// Only include profiles that have an assumable_role_id
		if credential.AssumableRoleID != "" {
			assumableRoles = append(assumableRoles, credential.AssumableRoleID)
		}
	}

	return assumableRoles
}

// GetProfileByRoleArn returns the profile name that has the given assumable_role_id
func (cr *CredentialReader) GetProfileByRoleArn(roleArn string) string {
	return cr.roleArnToProfile[roleArn]
}
