package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/reconcile.go")
	s := string(b)
	
	// Keep only the OIDC provider logic
	startStr := "func ReconcileAWSInfrastructure"
	endStr := "// 5. OIDC Provider"
	
	startIdx := strings.Index(s, startStr)
	endIdx := strings.Index(s, endStr)
	
	if startIdx != -1 && endIdx != -1 {
		header := s[:startIdx]
		
		newFunc := `func ReconcileAWSInfrastructure(runner terraform.Runner, projectName, awsRegion string) error {
	ui.PrintInfo("Checking global AWS account prerequisites...")

	stateList, err := runner.StateList()
	if err != nil {
		return fmt.Errorf("failed to read terraform state: %w", err)
	}

	stateMap := make(map[string]bool)
	for _, res := range stateList {
		stateMap[res] = true
	}

	// 1. OIDC Provider (Global singleton per AWS Account)
	if !stateMap["aws_iam_openid_connect_provider.github[0]"] && !stateMap["aws_iam_openid_connect_provider.github"] {
		// First get account ID
		accountIDOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
		if err == nil {
			accountID := strings.TrimSpace(string(accountIDOut))
			expectedArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/token.actions.githubusercontent.com", accountID)

			checkCmd := exec.Command("aws", "iam", "list-open-id-connect-providers", "--query", "OpenIDConnectProviderList[*].Arn", "--output", "text")
			if out, err := checkCmd.Output(); err == nil {
				if strings.Contains(string(out), "token.actions.githubusercontent.com") {
					ui.PrintSuccess("GitHub OIDC provider already exists")
					ui.PrintInfo("Importing OIDC provider into Terraform state...")
					if err := runner.Import("aws_iam_openid_connect_provider.github[0]", expectedArn); err != nil {
						runner.Import("aws_iam_openid_connect_provider.github", expectedArn)
					}
					ui.PrintSuccess("Imported into Terraform state")
					
					ui.PrintInfo("Ensuring OIDC Provider thumbprints are up-to-date...")
					exec.Command("aws", "iam", "update-open-id-connect-provider-thumbprint",
						"--open-id-connect-provider-arn", expectedArn,
						"--thumbprint-list", "6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd", "1b511abead59c6ce207077c0bf0e0043b1382612", "06d927fecd0a84aeba28aad1d808139470fe95c3", "ffffffffffffffffffffffffffffffffffffffff",
					).Run()
				}
			}
		}
	}

	return nil
}`
		
		os.WriteFile("internal/cli/reconcile.go", []byte(header+newFunc), 0644)
	}
}
