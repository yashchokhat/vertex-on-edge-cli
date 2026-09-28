package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yashchokhat/vertex-on-edge/internal/terraform/aws/ec2"
)

// RenderTemplates copies the embedded Terraform templates for the selected cloud provider
// and target into the specified project directory.
func RenderTemplates(projectDir, provider, target string, variables map[string]string) (string, error) {
	tfDir := filepath.Join(projectDir, ".vertex-on-edge", "terraform", provider, target)
	err := os.MkdirAll(tfDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create terraform directory: %w", err)
	}

	var templatesFS fs.FS

	// Select the embedded filesystem
	if provider == "aws" && target == "ec2" {
		templatesFS = ec2.Templates
	} else {
		return "", fmt.Errorf("unsupported provider/target combination: %s/%s", provider, target)
	}

	// Copy files
	entries, err := fs.ReadDir(templatesFS, ".")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := fs.ReadFile(templatesFS, entry.Name())
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
		}

		destPath := filepath.Join(tfDir, entry.Name())
		err = os.WriteFile(destPath, content, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
	}

	// Generate terraform.tfvars
	var tfvars string
	for k, v := range variables {
		// Very basic escaping for string variables
		tfvars += fmt.Sprintf("%s = \"%s\"\n", k, v)
	}

	err = os.WriteFile(filepath.Join(tfDir, "terraform.tfvars"), []byte(tfvars), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write tfvars: %w", err)
	}

	return tfDir, nil
}
