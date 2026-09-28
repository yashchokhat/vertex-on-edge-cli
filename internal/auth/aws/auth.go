package aws

import (
	"os"
	"os/exec"
)

type AWSAuthenticator struct {
	Identity string
}

func New() *AWSAuthenticator {
	return &AWSAuthenticator{}
}

func (a *AWSAuthenticator) IsAuthenticated() bool {
	// Simple check using aws sts get-caller-identity
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	err := cmd.Run()
	return err == nil
}

func (a *AWSAuthenticator) Authenticate() error {
	// For AWS, we assume they configure SSO or access keys manually via aws configure.
	// We can invoke aws configure if needed.
	cmd := exec.Command("aws", "configure")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *AWSAuthenticator) GetIdentifier() string {
	return "AWS Account"
}
