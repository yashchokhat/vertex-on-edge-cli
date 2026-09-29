package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/deploy.go")
	s := string(b)
	
	oldStr := `	appUrl, err := runner.Output("application_url")
	if err != nil {
		ui.PrintError("Failed to get Application URL from Terraform", err.Error())
	}

	ui.PrintInfo("Securing GitHub Actions environment...")
	secrets := map[string]string{
		"AWS_REGION":          awsRegion,
		"ECR_REPOSITORY_NAME": safeProjectName,
	}`

	newStr := `	appUrl, err := runner.Output("application_url")
	if err != nil {
		ui.PrintError("Failed to get Application URL from Terraform", err.Error())
	}
	
	instanceId, err := runner.Output("instance_id")
	if err != nil {
		// Non-fatal, just a warning if it doesn't exist yet
	}

	ui.PrintInfo("Securing GitHub Actions environment...")
	secrets := map[string]string{
		"AWS_REGION":          awsRegion,
		"ECR_REPOSITORY_NAME": safeProjectName,
	}
	
	if instanceId != "" {
		secrets["EC2_INSTANCE_ID"] = strings.TrimSpace(instanceId)
	}`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/cli/deploy.go", []byte(s), 0644)
}
