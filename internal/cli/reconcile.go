package cli

import (
	"github.com/yashchokhat/vertex-on-edge/internal/terraform"
)

// ReconcileAWSInfrastructure checks for existing AWS resources that Terraform
// expects to manage and imports them into the state file if they exist but are
// missing from the state. This prevents "resource already exists" errors on
// subsequent deploys.
func ReconcileAWSInfrastructure(runner terraform.Runner, projectName, awsRegion string) error {
	// No global singletons require importing at this time.
	// IAM users and policies are generated per-project.
	return nil
}
