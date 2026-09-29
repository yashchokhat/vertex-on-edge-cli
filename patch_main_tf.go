package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	// Add random_id properly
	if !strings.Contains(s, "resource \"random_id\" \"suffix\"") {
		s = strings.Replace(s, "provider \"aws\" {\n  region = var.aws_region\n}", "provider \"aws\" {\n  region = var.aws_region\n}\n\nresource \"random_id\" \"suffix\" {\n  byte_length = 3\n}", 1)
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

	// Ensure thumbprint list is correct for OIDC
	if !strings.Contains(s, "6938fd4d98bab03faadb97b34396831e3780aea1") {
		s = strings.Replace(s, "client_id_list  = [\"sts.amazonaws.com\"]", "client_id_list  = [\"sts.amazonaws.com\"]\n  thumbprint_list = [\"6938fd4d98bab03faadb97b34396831e3780aea1\", \"1c58a3a8518e8759bf075b76b750d4f2df264fcd\", \"1b511abead59c6ce207077c0bf0e0043b1382612\", \"06d927fecd0a84aeba28aad1d808139470fe95c3\", \"ffffffffffffffffffffffffffffffffffffffff\"]", 1)
	}

	// Patch wildcard trust policy
	oldTrust := `        Condition = {
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_repo}:*"
          }
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
        }`
        
    oldTrust2 := `        Condition = {
          StringLike = {
            "token.actions.githubusercontent.com:sub" = [
              "repo:${var.github_owner}/${var.github_repository}:*",
              "repo:${var.github_owner}@*/${var.github_repository}@*:*"
            ]
          }
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
        }`
        
    newTrust := `        Condition = {
          StringLike = {
            "token.actions.githubusercontent.com:sub" = [
              "repo:${var.github_owner}/*:*",
              "repo:${var.github_owner}@*/*:*"
            ]
          }
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
        }`

	if strings.Contains(s, oldTrust) {
		s = strings.Replace(s, oldTrust, newTrust, 1)
	} else if strings.Contains(s, oldTrust2) {
	    s = strings.Replace(s, oldTrust2, newTrust, 1)
	}

	// Add AmazonSSMManagedInstanceCore to EC2 role
	if !strings.Contains(s, "AmazonSSMManagedInstanceCore") {
		ssmPolicy := `
resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}
`
		s = s + ssmPolicy
	}

	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
