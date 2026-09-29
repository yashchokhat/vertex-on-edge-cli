package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	oldStr := `        Condition = {
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

	newStr := `        Condition = {
          StringLike = {
            "token.actions.githubusercontent.com:sub" = [
              "repo:${var.github_owner}/*:*"
            ]
          }
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
        }`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
