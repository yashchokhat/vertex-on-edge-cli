package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/github/actions.go")
	s := string(b)
	
	oldStr := `      - name: Trigger EC2 deployment
        run: |
          echo "Deployment triggered to EC2 via AWS Systems Manager"
          # aws ssm send-command ...`

	newStr := `      - name: Trigger EC2 deployment
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: ${{ secrets.ECR_REPOSITORY_NAME }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          aws ssm send-command \
            --instance-ids "${{ secrets.EC2_INSTANCE_ID }}" \
            --document-name "AWS-RunShellScript" \
            --parameters "commands=[
              \"aws ecr get-login-password --region ${{ secrets.AWS_REGION }} | docker login --username AWS --password-stdin $ECR_REGISTRY\",
              \"docker pull $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG\",
              \"docker stop app || true\",
              \"docker rm app || true\",
              \"docker run -d --name app -p 80:3000 -p 3000:3000 $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG\"
            ]"`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/github/actions.go", []byte(s), 0644)
}
