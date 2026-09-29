package main

import (
	"os"
	"strings"
)

func main() {
	// Patch outputs.tf
	bOut, _ := os.ReadFile("internal/terraform/aws/ec2/outputs.tf")
	sOut := string(bOut)
	idx := strings.Index(sOut, "output \"github_actions_role_arn\"")
	if idx != -1 {
		sOut = sOut[:idx]
	}
	os.WriteFile("internal/terraform/aws/ec2/outputs.tf", []byte(sOut), 0644)
	
	// Patch main.tf to remove github_actions_role
	bMain, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	sMain := string(bMain)
	
	startIdx := strings.Index(sMain, "resource \"aws_iam_role\" \"github_actions_role\"")
	if startIdx != -1 {
		sMain = sMain[:startIdx]
	}
	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(sMain), 0644)
}
