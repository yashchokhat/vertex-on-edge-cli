package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/terraform/aws/ec2/main.tf")
	s := string(b)
	
	oldStr := `# Attach ECR Read Policy to EC2 Role
resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}`

	newStr := `# Attach ECR Read Policy to EC2 Role
resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# Attach SSM Core Policy to EC2 Role so GitHub Actions can trigger deployments
resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/terraform/aws/ec2/main.tf", []byte(s), 0644)
}
