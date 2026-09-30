package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yashchokhat/vertex-on-edge/internal/ui"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Prints out diagnostic information for AWS OIDC and GitHub Actions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner()
		ui.PrintInfo("Gathering Vertex-on-Edge Diagnostic Info...")

		// 1. Get AWS Identity
		callerOut, _ := exec.Command("aws", "sts", "get-caller-identity").Output()
		ui.PrintInfo(fmt.Sprintf("AWS Caller Identity:\n%s", string(callerOut)))

		// 2. Get GitHub User & Repo
		projectPath, _ := os.Getwd()
		
		ghUserOut, _ := exec.Command("gh", "api", "user", "-q", ".login").Output()
		ghUser := strings.TrimSpace(string(ghUserOut))
		ui.PrintInfo(fmt.Sprintf("GitHub Authenticated User: %s", ghUser))

		ghRepoCmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
		ghRepoCmd.Dir = projectPath
		ghRepoOut, _ := ghRepoCmd.Output()
		ui.PrintInfo(fmt.Sprintf("Current GitHub Repo: %s", strings.TrimSpace(string(ghRepoOut))))

		// 3. Find the IAM Role name
		tfDir := projectPath + "/.vertex-on-edge/terraform/aws/ec2"
		
		if _, err := os.Stat(tfDir); err == nil {
			roleArnOut, err := exec.Command("terraform", "-chdir="+tfDir, "output", "-raw", "github_actions_role_arn").Output()
			if err == nil && len(roleArnOut) > 0 {
				roleArn := strings.TrimSpace(string(roleArnOut))
				ui.PrintSuccess(fmt.Sprintf("Found Role ARN: %s", roleArn))
				
				parts := strings.Split(roleArn, "/")
				if len(parts) > 1 {
					roleName := parts[1]
					roleOut, err := exec.Command("aws", "iam", "get-role", "--role-name", roleName).CombinedOutput()
					if err != nil {
						ui.PrintError("Failed to get role details", string(roleOut))
					} else {
						ui.PrintInfo(fmt.Sprintf("IAM Role Trust Policy:\n%s", string(roleOut)))
					}
				}
			} else {
				ui.PrintError("Could not get role ARN from terraform. Has it been deployed?", "")
			}
		} else {
			ui.PrintError(fmt.Sprintf("No terraform state found in this directory: %s", tfDir), "")
		}

		// 4. Get OIDC Provider info
		accountIDOut, _ := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text").Output()
		accountID := strings.TrimSpace(string(accountIDOut))
		oidcArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/token.actions.githubusercontent.com", accountID)
		
		oidcOut, err := exec.Command("aws", "iam", "get-open-id-connect-provider", "--open-id-connect-provider-arn", oidcArn).CombinedOutput()
		if err != nil {
			ui.PrintError("Failed to get OIDC provider", string(oidcOut))
		} else {
			ui.PrintInfo(fmt.Sprintf("OIDC Provider Info:\n%s", string(oidcOut)))
		}

		// 5. List GitHub Secrets
		secretsCmd := exec.Command("gh", "secret", "list")
		secretsCmd.Dir = projectPath
		secretsOut, _ := secretsCmd.Output()
		ui.PrintInfo(fmt.Sprintf("GitHub Secrets:\n%s", string(secretsOut)))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
