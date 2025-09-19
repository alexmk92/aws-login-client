#!/usr/bin/env zsh

# JJ AWS Login Helper
# Usage: aws-login <profile>
# Example: aws-login int

# If this is commented out, you will be prompted to select an authentication driver
# can either be manual or 1password
# export AWS_LOGIN_AUTH_DRIVER=1password

# ECR_REGISTRY is the ECR registry to login to if you want to login to it
# export ECR_REGISTRY=ACCOUNT_ID.dkr.ecr.eu-west-2.amazonaws.com

# Set AWS profile and run the Go application
export AWS_PROFILE=$1
# If you compile with go build -o aws-login, you can use the following line instead:
# and then move the compiled binary to /usr/local/bin and make it executable
# sudo mv aws-login /usr/local/bin
# sudo chmod +x /usr/local/bin/aws-login
# You can then use the following line instead:
# aws-login $1
go run main.go $1

# Load credentials from JSON file and clean up
if [ -f "/tmp/aws-session.json" ]; then
    export AWS_ACCESS_KEY_ID=$(jq -r '.AccessKeyId' /tmp/aws-session.json)
    export AWS_SECRET_ACCESS_KEY=$(jq -r '.SecretAccessKey' /tmp/aws-session.json)
    export AWS_SESSION_TOKEN=$(jq -r '.SessionToken' /tmp/aws-session.json)
    export AWS_PROFILE=$(jq -r '.ProfileName' /tmp/aws-session.json)

    # Cleanup to not expose the credentials to the shell
    rm /tmp/aws-session.json
fi
