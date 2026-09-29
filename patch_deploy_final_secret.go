package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	targetBlock := `	if instanceId != "" {
		secrets["EC2_INSTANCE_ID"] = strings.TrimSpace(instanceId)
	}`
	
	newBlock := `	if instanceId != "" {
		secrets["EC2_INSTANCE_ID"] = strings.TrimSpace(instanceId)
	}

	// Inject the global vertexOnEdge-cli IAM role ARN
	accountIDOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
	if err == nil {
		accountID := strings.TrimSpace(string(accountIDOut))
		if accountID != "" {
			secrets["AWS_ROLE_ARN"] = "arn:aws:iam::" + accountID + ":role/vertexOnEdge-cli"
		}
	}`
	
	s = strings.Replace(s, targetBlock, newBlock, 1)
	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
