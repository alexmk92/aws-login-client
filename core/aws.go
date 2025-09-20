package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/alexmk92/aws-login/core/types"
)

// AWSService handles all AWS-related operations
type AWSService struct {
	credentialReader *CredentialReader
	attemptECRLogin  bool
}

// Create a new AWS service instance, if we wanted this to be a singleton
// for a thread safe singleton, we could use the sync do once pattern
//
// In modern Go (since 1.18), it's recommended to use 'any' instead of 'interface{}' for an empty interface.
// below is an example of a thread-safe singleton pattern for AWSService.
//
// import "sync"
//
// var (
//
//	awsServiceInstance *AWSService
//	awsServiceOnce     sync.Once
//
// )
//
//	func GetAWSService() *AWSService {
//	 // this block will only be executed once synchronously across all goroutines
//	 // it will be executed in a thread safe manner (go routines will block each other until the block is executed)
//		awsServiceOnce.Do(func() {
//			awsServiceInstance = &AWSService{
//				credentialReader: NewCredentialReader(),
//			}
//		})
//
//		return awsServiceInstance
//	}
func NewAWSService(attemptECRLogin bool) *AWSService {
	credentialReader := NewCredentialReader()
	if err := credentialReader.LoadCredentialsFile(); err != nil {
		// This is a fatal error, we need to load the credentials file, it will send an os.Exit(1) signal
		log.Fatalf("Failed to load credentials file: %v", err)
	}

	return &AWSService{
		credentialReader: credentialReader,
		attemptECRLogin:  attemptECRLogin,
	}
}

// GetCredentials returns the credentials for a specific profile
// notice that we're returning a nil pointer if the credential is not found
// this is because we want to allow the caller to handle the error case gracefully
// if you don't handle the error case and simply return the nil, it would be easy
// to create a panic (crash) due to a nil pointer dereference.
//
// The issue with this approach is we're returning a pointer to the credential,
// which means we're passing around references to the credential, and therefore
// any modifications to the credential will be reflected in the credential readers
// internal state.
//
// We could choose to return a copy instead, this would result in sizeof(types.StaticCredential) bytes
// of memory being allocated for each call to GetCredentials, which is probably fine in the
// case of our application.  If you return a value instead of a * you cannot return nil
// instead you'd need to return a zero value for the type, such as types.StaticCredential{}
func (s *AWSService) GetCredentials(profile string) (*types.StaticCredential, error) {
	if s.credentialReader == nil {
		return nil, fmt.Errorf("credential reader not initialized")
	}

	credential, exists := s.credentialReader.GetCredential(profile)
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found in credentials", profile)
	}

	return &credential, nil
}

func (s *AWSService) GetMFASerial(profile string) (string, error) {
	credentials, err := s.GetCredentials(profile)
	if err != nil {
		return "", err
	}

	if credentials.MfaSerial == "" {
		return "", fmt.Errorf("MFA serial not configured for profile '%s'", profile)
	}

	return credentials.MfaSerial, nil
}

func (s *AWSService) GetMFACode(authDriver types.Driver) (string, error) {
	if authDriver == nil {
		return "", fmt.Errorf("auth driver not initialized")
	}

	if !authDriver.YieldsMFACode() {
		return "", fmt.Errorf("auth driver does not yield MFA code")
	}

	return authDriver.GetMFACode()
}

// GetValidProfiles returns a list of all valid profile names from the AWS service
func (s *AWSService) GetValidProfiles() []string {
	if s.credentialReader == nil {
		return []string{}
	}

	return s.credentialReader.GetValidProfiles()
}

// GetAssumableRoles returns the list of roles that can be assumed for a profile
func (s *AWSService) GetAssumableRoles(profile string) []string {
	if s.credentialReader == nil {
		return []string{}
	}

	return s.credentialReader.GetAssumableRoles(profile)
}

// GetAssumedProfileName returns the profile name that has the given assumable_role_id
func (aws *AWSService) GetAssumedProfileName(roleArn string) string {
	if aws.credentialReader == nil {
		return ""
	}

	return aws.credentialReader.GetProfileByRoleArn(roleArn)
}

// ValidateMFACode checks if the MFA code is 6 digits
func (s *AWSService) ValidateMFACode(code string) bool {
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

// GetSessionToken gets temporary AWS credentials using provided MFA code
// all types.Credentials yielded by getSessionTokenInternal are set in the process
// environment variables, so we don't need to return them.
//
// This means the AWS service can govern who the current active session belongs
// to without a user accidently changing the profile.
func (s *AWSService) GetSessionToken(profile, mfaCode string) (bool, error) {
	mfaSerial, err := s.GetMFASerial(profile)
	if err != nil {
		return false, err
	}

	// Get session token
	cmd := exec.Command("aws", "sts", "get-session-token",
		"--duration", "86400",
		"--serial-number", mfaSerial,
		"--token-code", mfaCode,
		"--profile", profile)

	// Debug print removed to avoid interfering with Bubble Tea rendering

	output, err := cmd.Output()

	if err != nil {
		return false, fmt.Errorf("failed to get AWS session token: %w", err)
	}

	var stsResponse types.STSResponse
	if err := json.Unmarshal(output, &stsResponse); err != nil {
		return false, fmt.Errorf("failed to parse STS response: %w", err)
	}

	return s.persistCredentials(&stsResponse.Credentials, profile)
}

// LoginToECR performs Docker login to ECR using temporary credentials
func (s *AWSService) LoginToECR() error {
	if !s.attemptECRLogin {
		return fmt.Errorf("attempt to login to ECR is disabled")
	}

	credentials, err := s.GetCredentials(os.Getenv("AWS_PROFILE"))
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	// Get ECR login password using temporary credentials
	passwordCmd := exec.Command("aws", "ecr", "get-login-password", "--region", "eu-west-2")
	password, err := passwordCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ECR login password: %w", err)
	}

	// Ensure we have an account ID, AccountID can be optional in the credentials file, but the
	// user is required to specify the full RoleARN for the assumable role if we're using that instead.
	accountID := credentials.AccountID
	if accountID == "" {
		accountID = credentials.AssumableRoleID
		// Role ARN = arn:aws:iam::ACCOUNT:role/ROLE_NAME
		// we want to extract the ACCOUNT ID
		accountID = strings.Split(accountID, ":")[4]
	}

	// Docker login
	dockerCmd := exec.Command("docker", "login",
		"--username", "AWS",
		"--password-stdin",
		fmt.Sprintf("%s.dkr.ecr.eu-west-2.amazonaws.com", accountID))
	dockerCmd.Stdin = strings.NewReader(string(password))

	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("failed to login to ECR: %w", err)
	}

	return nil
}

// AssumeRole assumes a role using the current session credentials
func (s *AWSService) AssumeRole(profile string, roleArn string) (bool, error) {
	// Call assume-role
	cmd := exec.Command("aws", "sts", "assume-role",
		"--role-arn", strings.TrimSpace(roleArn),
		"--role-session-name", "aws-login-session")

	output, err := cmd.Output()

	// Also check if there's stderr output
	if exitError, ok := err.(*exec.ExitError); ok {
		return false, fmt.Errorf("failed to assume role %s: %w", roleArn, exitError)
	}

	if err != nil {
		return false, fmt.Errorf("failed to assume role %s: %w", roleArn, err)
	}

	var assumeResponse types.AssumeRoleResponse
	if err := json.Unmarshal(output, &assumeResponse); err != nil {
		return false, fmt.Errorf("failed to parse assume-role response: %w", err)
	}

	return s.persistCredentials(&assumeResponse.Credentials, profile)
}

func (s *AWSService) persistCredentials(credentials *types.Credentials, profile string) (bool, error) {
	// Persist the credentials to the environment for the remainder
	// of this programs execution.
	os.Setenv("AWS_ACCESS_KEY_ID", credentials.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", credentials.SessionToken)
	os.Setenv("AWS_PROFILE", profile)

	token := &types.Credentials{
		AccessKeyId:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Profile:         profile,
		Expiration:      credentials.Expiration,
	}

	// Write the credentials to a JSON file, we don't return credentials
	// because we don't want other parts of the program to have access to
	// them as they are sensitive and should not be tampered with.
	_, err := s.writeToJSONFile(token, "/tmp/aws-session.json")
	if err != nil {
		return false, fmt.Errorf("failed to write credentials to JSON file: %w", err)
	}

	return true, nil
}

// Private helper function to write credentials to a JSON file
// this is used to persist the credentials to a file for the remainder
// of this programs execution. The bash script will then source this
// file to set the environment variables and then remove the file
// once the credentials are no longer needed.  We write to /tmp
// as the files will be cleaned up by the OS on reboot.
func (s *AWSService) writeToJSONFile(credentials *types.Credentials, filePath string) (string, error) {
	// Create JSON structure
	jsonData := map[string]interface{}{
		"Version":         1,
		"AccessKeyId":     credentials.AccessKeyId,
		"SecretAccessKey": credentials.SecretAccessKey,
		"SessionToken":    credentials.SessionToken,
		"Expiration":      credentials.Expiration,
		"ProfileName":     credentials.Profile,
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file: %w", err)
	}

	return filePath, nil
}
