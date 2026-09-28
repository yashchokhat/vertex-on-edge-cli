package platform

import (
	"os"
	"path/filepath"
)

// DetectExistingConfig checks if the project has already been configured for Vertex-on-Edge.
// It returns true, providerID, targetID, tfDir if found.
func DetectExistingConfig(projectPath string) (bool, string, string, string) {
	terraformBase := filepath.Join(projectPath, ".vertex-on-edge", "terraform")
	if _, err := os.Stat(terraformBase); os.IsNotExist(err) {
		return false, "", "", ""
	}

	// Try to find the provider and target
	// Typically: .vertex-on-edge/terraform/<provider>/<target>
	providers, err := os.ReadDir(terraformBase)
	if err != nil || len(providers) == 0 {
		return false, "", "", ""
	}

	for _, p := range providers {
		if p.IsDir() {
			providerName := p.Name()
			targets, err := os.ReadDir(filepath.Join(terraformBase, providerName))
			if err == nil && len(targets) > 0 {
				for _, t := range targets {
					if t.IsDir() {
						targetName := t.Name()
						tfDir := filepath.Join(terraformBase, providerName, targetName)
						// Check if main.tf exists
						if _, err := os.Stat(filepath.Join(tfDir, "main.tf")); err == nil {
							return true, providerName, targetName, tfDir
						}
					}
				}
			}
		}
	}

	return false, "", "", ""
}
