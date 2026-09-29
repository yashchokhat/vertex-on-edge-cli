package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	// Change the github_actions_role to use a fixed name
	oldName := `name = "${var.project_name}-github-actions-role-${random_id.suffix.hex}"`
	newName := `name = "vertexOnEdge-cli"`
	s = strings.Replace(s, oldName, newName, 1)

	// Change the inline policy to allow pushing to ANY ECR repository
	oldEcr := `Resource = aws_ecr_repository.app.arn`
	newEcr := `Resource = "*"`
	s = strings.Replace(s, oldEcr, newEcr, 1)

	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
