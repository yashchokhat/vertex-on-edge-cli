package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	// Create vertexOnEdge-cli role globally before rendering templates
	insertCode := `		ui.PrintInfo("Ensuring global GitHub Actions IAM role exists...")
		// Trust policy for vertexOnEdge-cli
		trustPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Federated\":\"arn:aws:iam::\" + strings.TrimSpace(string(accountIDOut)) + \":oidc-provider/token.actions.githubusercontent.com\"},\"Action\":\"sts:AssumeRoleWithWebIdentity\",\"Condition\":{\"StringEquals\":{\"token.actions.githubusercontent.com:aud\":\"sts.amazonaws.com\"},\"StringLike\":{\"token.actions.githubusercontent.com:sub\":[\"repo:\" + githubOwner + \"/*:*\",\"repo:\" + githubOwner + \"@*/*:*\"]}}}]}"
		
		accountIDOut, _ := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
		trustPolicy = strings.ReplaceAll(trustPolicy, "strings.TrimSpace(string(accountIDOut))", strings.TrimSpace(string(accountIDOut)))
		
		// Create role if not exists
		checkRoleCmd := exec.Command("aws", "iam", "get-role", "--role-name", "vertexOnEdge-cli")
		if err := checkRoleCmd.Run(); err != nil {
			createRoleCmd := exec.Command("aws", "iam", "create-role", "--role-name", "vertexOnEdge-cli", "--assume-role-policy-document", trustPolicy)
			createRoleCmd.Run()
			
			// Attach policies
			exec.Command("aws", "iam", "attach-role-policy", "--role-name", "vertexOnEdge-cli", "--policy-arn", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser").Run()
			exec.Command("aws", "iam", "attach-role-policy", "--role-name", "vertexOnEdge-cli", "--policy-arn", "arn:aws:iam::aws:policy/AmazonSSMFullAccess").Run()
		} else {
			// Update trust policy just in case githubOwner changed
			updateTrustCmd := exec.Command("aws", "iam", "update-assume-role-policy", "--role-name", "vertexOnEdge-cli", "--policy-document", trustPolicy)
			updateTrustCmd.Run()
		}
`

	// This is a bit too hacky to inject with strings.Replace reliably into deploy.go
	os.WriteFile("internal/cli/deploy.go.patch", []byte(insertCode), 0644)
}
