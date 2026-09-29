package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	// Add random_id resource at the top
	if !strings.Contains(s, "resource \"random_id\" \"suffix\"") {
		s = strings.Replace(s, "provider \"aws\" {", "provider \"aws\" {\n}\n\nresource \"random_id\" \"suffix\" {\n  byte_length = 3\n", 1)
	}

	// Patch ECR
	s = strings.ReplaceAll(s, `name                 = var.project_name`, `name                 = "${var.project_name}-${random_id.suffix.hex}"`)
	
	// Patch Key Pair
	s = strings.ReplaceAll(s, `key_name   = "${var.project_name}-key"`, `key_name   = "${var.project_name}-key-${random_id.suffix.hex}"`)
	
	// Patch EC2 Role
	s = strings.ReplaceAll(s, `name = "${var.project_name}-ec2-role"`, `name = "${var.project_name}-ec2-role-${random_id.suffix.hex}"`)
	
	// Patch Instance Profile
	s = strings.ReplaceAll(s, `name = "${var.project_name}-profile"`, `name = "${var.project_name}-profile-${random_id.suffix.hex}"`)
	
	// Patch GitHub Actions Role
	s = strings.ReplaceAll(s, `name = "${var.project_name}-github-actions-role"`, `name = "${var.project_name}-github-actions-role-${random_id.suffix.hex}"`)

	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
