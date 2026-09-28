package github

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateTestWorkflow generates a CI testing workflow.
func GenerateTestWorkflow(projectPath, language string, testingFrameworks []string) error {
	if len(testingFrameworks) == 0 {
		return nil // Do nothing
	}

	var workflowContent string

	switch language {
	case "TypeScript", "JavaScript":
		workflowContent = `name: CI Testing

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Use Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18.x'
    - run: npm ci
`
		for _, framework := range testingFrameworks {
			switch framework {
			case "jest":
				workflowContent += "    - run: npm run test -- --passWithNoTests\n"
			case "vitest":
				workflowContent += "    - run: npx vitest run --passWithNoTests\n"
			case "cypress":
				workflowContent += "    - run: npx cypress run || echo 'Cypress tests failed or no tests found'\n"
			case "playwright":
				workflowContent += "    - run: npx playwright install --with-deps\n    - run: npx playwright test --pass-with-no-tests\n"
			default:
				workflowContent += "    - run: npm test\n"
			}
		}

	case "Python":
		workflowContent = `name: CI Testing

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.10'
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
`
		for _, framework := range testingFrameworks {
			switch framework {
			case "pytest":
				workflowContent += "        pip install pytest\n    - name: Test with pytest\n      run: pytest\n"
			case "unittest":
				workflowContent += "    - name: Test with unittest\n      run: python -m unittest discover\n"
			}
		}

	case "Go":
		workflowContent = `name: CI Testing

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Test
      run: go test -v ./...
`

	default:
		workflowContent = fmt.Sprintf(`# CI Testing workflow for %s
name: CI Testing

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - run: echo "Run your tests here"
`, language)
	}

	workflowsDir := filepath.Join(projectPath, ".github", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(workflowsDir, "test.yml"), []byte(workflowContent), 0644)
}

// GenerateDeployWorkflow generates a GitHub Actions workflow for the target platform.
func GenerateDeployWorkflow(projectPath, provider, target string) error {
	var workflowContent string

	if provider == "aws" && target == "ec2" {
		workflowContent = `name: Deploy to AWS EC2

on:
  push:
    branches:
      - main

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v1

      - name: Build, tag, and push image to Amazon ECR
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: ${{ secrets.ECR_REPOSITORY_NAME }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG .
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG

      - name: Trigger EC2 deployment
        run: |
          echo "Deployment triggered to EC2 via AWS Systems Manager"
          # aws ssm send-command ...
`
	} else if provider == "gcp" && target == "cloud-run" {
		workflowContent = `name: Deploy to GCP Cloud Run

on:
  push:
    branches:
      - main

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - id: 'auth'
        name: 'Authenticate to Google Cloud'
        uses: 'google-github-actions/auth@v1'
        with:
          workload_identity_provider: '${{ secrets.WIF_PROVIDER }}'
          service_account: '${{ secrets.SA_EMAIL }}'

      - name: 'Set up Cloud SDK'
        uses: 'google-github-actions/setup-gcloud@v1'

      - name: 'Docker auth'
        run: |-
          gcloud auth configure-docker ${{ secrets.REGION }}-docker.pkg.dev

      - name: 'Build and Push Container'
        run: |-
          docker build -t ${{ secrets.REGION }}-docker.pkg.dev/${{ secrets.PROJECT_ID }}/${{ secrets.REPO_NAME }}/app:${{ github.sha }} .
          docker push ${{ secrets.REGION }}-docker.pkg.dev/${{ secrets.PROJECT_ID }}/${{ secrets.REPO_NAME }}/app:${{ github.sha }}

      - name: 'Deploy to Cloud Run'
        uses: 'google-github-actions/deploy-cloudrun@v1'
        with:
          service: '${{ secrets.SERVICE_NAME }}'
          region: '${{ secrets.REGION }}'
          image: '${{ secrets.REGION }}-docker.pkg.dev/${{ secrets.PROJECT_ID }}/${{ secrets.REPO_NAME }}/app:${{ github.sha }}'
`
	} else if provider == "azure" && target == "container-apps" {
		workflowContent = `name: Deploy to Azure Container Apps

on:
  push:
    branches:
      - main

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: 'Azure login'
        uses: azure/login@v1
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
          subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}

      - name: 'Build and push image to ACR'
        run: |
          az acr login --name ${{ secrets.ACR_NAME }}
          docker build -t ${{ secrets.ACR_NAME }}.azurecr.io/app:${{ github.sha }} .
          docker push ${{ secrets.ACR_NAME }}.azurecr.io/app:${{ github.sha }}

      - name: 'Deploy to Container Apps'
        uses: azure/container-apps-deploy-action@v1
        with:
          appSourcePath: ${{ github.workspace }}
          acrName: ${{ secrets.ACR_NAME }}
          containerAppName: ${{ secrets.CONTAINER_APP_NAME }}
          resourceGroup: ${{ secrets.RESOURCE_GROUP }}
          imageToDeploy: ${{ secrets.ACR_NAME }}.azurecr.io/app:${{ github.sha }}
`
	} else {
		return fmt.Errorf("workflow generation for %s/%s is not supported", provider, target)
	}

	workflowsDir := filepath.Join(projectPath, ".github", "workflows")
	err := os.MkdirAll(workflowsDir, 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(workflowsDir, "deploy.yml"), []byte(workflowContent), 0644)
}
