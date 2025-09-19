# AWS Login

Secure AWS login with MFA support and 1Password integration.

## Features

- 🔐 Manual MFA or 1Password CLI authentication
- 🎨 Modern terminal UI
- 🚀 Automatic ECR login
- 🐚 Zsh shell integration

## Setup

1. **Build**: `go mod tidy && go build`

2. **Add to `~/.zshrc`**:
```bash
aws-login() {
    # If you wish to be auto logged into ECR, set this variable!
    # export ECR_REGISTRY=ACCOUNT_ID.dkr.ecr.eu-west-2.amazonaws.com

    go run main.go $1
    if [ -f "/tmp/aws-session.json" ]; then
        export AWS_ACCESS_KEY_ID=$(jq -r '.AccessKeyId' /tmp/aws-session.json)
        export AWS_SECRET_ACCESS_KEY=$(jq -r '.SecretAccessKey' /tmp/aws-session.json)
        export AWS_SESSION_TOKEN=$(jq -r '.SessionToken' /tmp/aws-session.json)
        rm /tmp/aws-session.json
    fi
}
```

3. **Configure AWS profiles** in `~/.aws/credentials`:
```ini
[profile-name]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
mfa_serial = arn:aws:iam::ACCOUNT:mfa/USERNAME
aws_account_id = 000000000000
assumable_role_id = arn:aws:iam::ACCOUNT:role/ROLE_NAME
vault_key = AWS MFA profile-name
```

**Optional fields:**
- `vault_key`: 1Password vault item name for automatic MFA retrieval
- `assumable_role_id`: IAM role ARN for cross-account access

## Usage

```bash
aws-login profile-name
```

## Requirements

- Go 1.25+
- AWS CLI configured
- MFA device
- jq
- 1Password CLI (optional)
