package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/yashchokhat/vertex-on-edge/internal/terraform"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

// ReconcileAWSInfrastructure checks for existing AWS resources that Terraform
// expects to manage and imports them into the state file if they exist but are
// missing from the state. This prevents "resource already exists" errors on
// subsequent deploys.
//
// Currently handles:
//   - GitHub OIDC provider (global singleton per AWS account)
func ReconcileAWSInfrastructure(runner terraform.Runner, projectName, awsRegion string) error {
	ui.PrintInfo("Checking global AWS account prerequisites...")

	stateList, err := runner.StateList()
	if err != nil {
		// An empty or missing state file is not an error -- it just means
		// this is a fresh deployment.
		return nil
	}

	stateMap := make(map[string]bool)
	for _, res := range stateList {
		stateMap[res] = true
	}

	// OIDC Provider: this is a global singleton per AWS account. If it already
	// exists (from a previous Vertex-on-Edge deploy, or from manual setup), we
	// need to import it so Terraform does not try to create a duplicate.
	if !stateMap["aws_iam_openid_connect_provider.github[0]"] && !stateMap["aws_iam_openid_connect_provider.github"] {
		accountIDOut, err := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
		if err != nil {
			// If we cannot determine the account ID, skip reconciliation.
			// Terraform will attempt to create the provider during apply.
			return nil
		}

		accountID := strings.TrimSpace(string(accountIDOut))
		expectedArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/token.actions.githubusercontent.com", accountID)

		checkCmd := exec.Command("aws", "iam", "list-open-id-connect-providers", "--query", "OpenIDConnectProviderList[*].Arn", "--output", "text")
		out, err := checkCmd.Output()
		if err != nil {
			return nil
		}

		if strings.Contains(string(out), "token.actions.githubusercontent.com") {
			ui.PrintSuccess("GitHub OIDC provider already exists in this AWS account")
			ui.PrintInfo("Importing OIDC provider into Terraform state...")

			// Try both the indexed and non-indexed resource addresses since
			// the resource uses count = var.create_oidc_provider ? 1 : 0.
			if err := runner.Import("aws_iam_openid_connect_provider.github[0]", expectedArn); err != nil {
				runner.Import("aws_iam_openid_connect_provider.github", expectedArn)
			}
			ui.PrintSuccess("Imported into Terraform state")

			// Update thumbprints to current known-good values.
			ui.PrintInfo("Ensuring OIDC Provider thumbprints are up-to-date...")
			exec.Command("aws", "iam", "update-open-id-connect-provider-thumbprint",
				"--open-id-connect-provider-arn", expectedArn,
				"--thumbprint-list",
				"6938fd4d98bab03faadb97b34396831e3780aea1",
				"1c58a3a8518e8759bf075b76b750d4f2df264fcd",
				"1b511abead59c6ce207077c0bf0e0043b1382612",
				"06d927fecd0a84aeba28aad1d808139470fe95c3",
				"ffffffffffffffffffffffffffffffffffffffff",
			).Run()
		}
	}

	return nil
}