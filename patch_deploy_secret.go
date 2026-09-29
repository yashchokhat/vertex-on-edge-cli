package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	// Remove the terraform runner output logic for roleArn
	oldOutputStr := `	roleArn, _ := runner.Output("github_actions_role_arn")
	if roleArn != "" {
		secrets["AWS_ROLE_ARN"] = strings.TrimSpace(roleArn)
	}`
	
	newOutputStr := `	// Use the global vertexOnEdge-cli IAM role as requested
	accountIDOut, _ := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
	accountID := strings.TrimSpace(string(accountIDOut))
	if accountID != "" {
		secrets["AWS_ROLE_ARN"] = "arn:aws:iam::" + accountID + ":role/vertexOnEdge-cli"
	}`
	
	s = strings.Replace(s, oldOutputStr, newOutputStr, 1)

	// We also need to fix Fast Deploy path where it validates the trust policy
	oldFastDeployValidate := `		// Parse the true role name directly from Terraform output to prevent Fast Deploy desyncs
		var roleName string
		parts := strings.Split(roleArn, "/")
		if len(parts) > 1 {
			roleName = parts[len(parts)-1]
		} else {
			roleName = safeProjectName + "-github-actions-role"
		}

		// Validate IAM Role Trust Policy
		trustCmd := exec.Command("aws", "iam", "get-role", "--role-name", roleName, "--query", "Role.AssumeRolePolicyDocument", "--output", "json")
		if out, err := trustCmd.Output(); err == nil {
			if !strings.Contains(string(out), "token.actions.githubusercontent.com") {
				ui.PrintError("OIDC_ROLE_TRUST_INVALID", fmt.Sprintf("The IAM role '%s' does not correctly reference the GitHub OIDC provider.", roleName))
				return nil
			}
		} else {
			ui.PrintError("OIDC_ROLE_MISSING", "Failed to retrieve IAM role trust policy: " + err.Error())
			return nil
		}`
		
	newFastDeployValidate := `		// Validate global IAM Role Trust Policy
		roleName := "vertexOnEdge-cli"
		trustCmd := exec.Command("aws", "iam", "get-role", "--role-name", roleName, "--query", "Role.AssumeRolePolicyDocument", "--output", "json")
		if out, err := trustCmd.Output(); err == nil {
			if !strings.Contains(string(out), "token.actions.githubusercontent.com") {
				ui.PrintError("OIDC_ROLE_TRUST_INVALID", fmt.Sprintf("The IAM role '%s' does not correctly reference the GitHub OIDC provider.", roleName))
				return nil
			}
		} else {
			ui.PrintError("OIDC_ROLE_MISSING", "Failed to retrieve global IAM role trust policy (vertexOnEdge-cli): " + err.Error())
			return nil
		}`
		
	s = strings.Replace(s, oldFastDeployValidate, newFastDeployValidate, 1)

	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
