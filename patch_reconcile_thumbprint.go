package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/reconcile.go")
	s := string(b)
	
	oldStr := `					if err := runner.Import("aws_iam_openid_connect_provider.github[0]", expectedArn); err != nil {
						// Fallback if count wasn't parsed properly
						runner.Import("aws_iam_openid_connect_provider.github", expectedArn)
					}
					ui.PrintSuccess("Imported into Terraform state")`

	newStr := `					if err := runner.Import("aws_iam_openid_connect_provider.github[0]", expectedArn); err != nil {
						// Fallback if count wasn't parsed properly
						runner.Import("aws_iam_openid_connect_provider.github", expectedArn)
					}
					ui.PrintSuccess("Imported into Terraform state")
					
					// Force update the thumbprints dynamically via AWS CLI to guarantee they are correct,
					// because Terraform often ignores thumbprint drift on imported resources.
					ui.PrintInfo("Ensuring OIDC Provider thumbprints are up-to-date...")
					exec.Command("aws", "iam", "update-open-id-connect-provider-thumbprint",
						"--open-id-connect-provider-arn", expectedArn,
						"--thumbprint-list", "6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd", "1b511abead59c6ce207077c0bf0e0043b1382612", "06d927fecd0a84aeba28aad1d808139470fe95c3", "ffffffffffffffffffffffffffffffffffffffff",
					).Run()`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/cli/reconcile.go", []byte(s), 0644)
}
