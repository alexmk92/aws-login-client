package core

import (
	"strings"
	"testing"
)

func TestAWSService_GetMFASerial(t *testing.T) {
	// Create a mock credential reader for testing
	cr := NewCredentialReader()

	// Load test credentials
	credentialsContent := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
mfa_serial = arn:aws:iam::123456789012:mfa/user

[int]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
mfa_serial = arn:aws:iam::123456789012:mfa/int-user

[no-mfa]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE2
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY2
# mfa_serial is missing`

	// Clear any existing credentials from previous tests
	cr.clearCredentials()
	err := cr.loadCredentialsFromContent(credentialsContent)
	if err != nil {
		t.Fatalf("Failed to load test credentials: %v", err)
	}

	// Create AWS service with mock credential reader
	awsService := &AWSService{
		credentialReader: cr,
	}

	tests := []struct {
		name           string
		profile        string
		expectedSerial string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "valid profile with MFA serial",
			profile:        "default",
			expectedSerial: "arn:aws:iam::123456789012:mfa/user",
			expectError:    false,
		},
		{
			name:           "valid profile with MFA serial - int",
			profile:        "int",
			expectedSerial: "arn:aws:iam::123456789012:mfa/int-user",
			expectError:    false,
		},
		{
			name:          "profile without MFA serial",
			profile:       "no-mfa",
			expectError:   true,
			errorContains: "MFA serial not configured for profile 'no-mfa'",
		},
		{
			name:          "nonexistent profile",
			profile:       "nonexistent",
			expectError:   true,
			errorContains: "profile 'nonexistent' not found in credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mfaSerial, err := awsService.GetMFASerial(tt.profile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if mfaSerial != tt.expectedSerial {
					t.Errorf("Expected MFA serial '%s', got '%s'", tt.expectedSerial, mfaSerial)
				}
			}
		})
	}
}

func TestAWSService_ValidateMFACode(t *testing.T) {
	// Create AWS service with mock credential reader to avoid file loading
	awsService := &AWSService{
		credentialReader: NewCredentialReader(),
	}

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "valid 6-digit code",
			code:     "123456",
			expected: true,
		},
		{
			name:     "valid 6-digit code with zeros",
			code:     "000000",
			expected: true,
		},
		{
			name:     "invalid - too short",
			code:     "12345",
			expected: false,
		},
		{
			name:     "invalid - too long",
			code:     "1234567",
			expected: false,
		},
		{
			name:     "invalid - contains letters",
			code:     "12345a",
			expected: false,
		},
		{
			name:     "invalid - contains special characters",
			code:     "123-45",
			expected: false,
		},
		{
			name:     "invalid - empty string",
			code:     "",
			expected: false,
		},
		{
			name:     "invalid - spaces",
			code:     "123 45",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := awsService.ValidateMFACode(tt.code)
			if result != tt.expected {
				t.Errorf("ValidateMFACode(%s) = %v, expected %v", tt.code, result, tt.expected)
			}
		})
	}
}
