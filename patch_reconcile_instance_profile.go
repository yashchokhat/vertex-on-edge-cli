package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/cli/reconcile.go")
	s := string(b)
	
	insertBlock := `	// 6. IAM Instance Profile
	profileName := projectName + "-profile"
	if !stateMap["aws_iam_instance_profile.ec2_profile"] {
		out, err := exec.Command("aws", "iam", "get-instance-profile", "--instance-profile-name", profileName, "--query", "InstanceProfile.InstanceProfileName", "--output", "text").Output()
		if err == nil && strings.TrimSpace(string(out)) == profileName {
			ui.PrintSuccess(fmt.Sprintf("IAM instance profile '%s' already exists", profileName))
			ui.PrintInfo("Importing IAM instance profile into Terraform state...")
			if err := runner.Import("aws_iam_instance_profile.ec2_profile", profileName); err != nil {
				// Don't fail the entire deployment, just print error
				ui.PrintError("Failed to import instance profile", err.Error())
			} else {
				ui.PrintSuccess("Imported into Terraform state")
			}
		}
	}
`

	// Insert before "return nil"
	s = strings.Replace(s, "\treturn nil\n}", insertBlock+"\treturn nil\n}", 1)
	os.WriteFile("internal/cli/reconcile.go", []byte(s), 0644)
}
