package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	oldBlock := `	roleArn, err := runner.Output("github_actions_role_arn")
	if err != nil {
		ui.PrintError("Failed to get IAM Role ARN from Terraform", err.Error())
	}`
	
	s = strings.Replace(s, oldBlock, "", 1)
	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
