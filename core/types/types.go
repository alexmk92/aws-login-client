package types

// When I'm designing packages, I like to keep the types in a separate file from the main code.
// the only types that should be in the main code are the ones that correspond to the service,
// definition, such as the AWSService sruct in @aws.go
//
// This allows us to import the types into other packages without encountering circular dependencies.
// if types or interfaces are private they should go into the package that requires them, go
// controls visibility based on the first letter of the type or interface name. So even though
// AWSService is in the /core package, it wouldn't have access to the 'test' struct thats commented out in this file.
//
// Only files in the "/core/types" package can access the 'test' struct.
// type test struct {
// 	Test string
// }
//
// It's also worth noting that packages cannot specify longer paths, they are always a single word.
// if you need to attach methods to a type, you can not extend the type outside of the package it is
// defined in.

// This holds the final status for the auth flow, it is used
// to display the result to the user.
type AuthFlowResult struct {
	User    string
	ECRAuth bool
}

// Credentials represents AWS temporary credentials
// go allows us to define how json is marshalled and unmarshalled
// it allows us to selectively omit fields from the json marshalling
//
// You can serialize a struct to json by using the json.Marshal function
// i.e. jsonBytes, err := json.Marshal(myStruct)
type Credentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
	Profile         string `json:"Profile"`
}

// AssumeRoleResponse represents the response from aws sts assume-role
// serialization doesn't just apply to primitive types, it applies
// to nested structs as well, in this example we are serializing the
// Credentials struct to have the json key "Credentials" whos
// values will be derived from the Credentials struct rules above.
type AssumeRoleResponse struct {
	Credentials     Credentials `json:"Credentials"`
	AssumedRoleUser struct {
		AssumedRoleId string `json:"AssumedRoleId"`
		Arn           string `json:"Arn"`
	} `json:"AssumedRoleUser"`
}

// STSResponse represents the AWS STS get-session-token response
type STSResponse struct {
	Credentials Credentials `json:"Credentials"`
}

// StaticCredential represents a static AWS credential from the credentials file
// including our custom fields added for this project (AssumeableRoleID and VaultKey)
type StaticCredential struct {
	ProfileName     string
	AccessKey       string
	AccessSecret    string
	AccountID       string
	MfaSerial       string
	AssumableRoleID string // ARN of the role that can be assumed by this profile
	VaultKey        string // Key in the 1Password vault for this profile (or whatever the password vault is)
}

// Driver defines the interface for authentication drivers
type Driver interface {
	GetToken() (string, error)
	Name() string
	YieldsMFACode() bool // If this is a password vault or something similar, we can yield a token to the caller
	GetMFACode() (string, error)
	IsInstalled() bool
}
