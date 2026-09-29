Vertex-on-Edge
Deploy your application to your own cloud with a single CLI.
Vertex-on-Edge is an open-source Go-based DevOps automation CLI that helps developers deploy web applications without manually configuring Docker, CI/CD, Terraform, GitHub Actions, or cloud infrastructure.
The CLI detects the project's technology stack, prepares deployment files, provisions cloud infrastructure using Terraform, configures GitHub Actions, and establishes an automated CI/CD pipeline.
Architecture
┌──────────────────┐
│   Your Project   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Vertex-on-Edge   │
│      CLI         │
└────────┬─────────┘
         │
    ┌────┴─────┐
    ▼          ▼
Project      Template
Detection    Engine
    │          │
    └────┬─────┘
         ▼
┌──────────────────┐
│ Docker + GitHub  │
│ Actions Workflow │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│    Terraform     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│   AWS / Cloud    │
│       EC2        │
└────────┬─────────┘
         │
         ▼
    Live Application
Features
Framework and technology-stack detection
Interactive project directory picker
Predefined Dockerfile templates
Automated Docker image configuration
Predefined GitHub Actions workflows
Build and test pipeline generation
Terraform-based infrastructure provisioning
AWS EC2 deployment
Amazon ECR container registry
GitHub Actions OIDC authentication
Automatic IAM role configuration
GitHub repository integration
GitHub Actions secret configuration
Automated CI/CD after initial deployment
Custom domain/DNS-ready deployment
Written in Go
No hosted Vertex-on-Edge control plane
Deployment Workflow
Developer
    │
    │ vertex-on-edge deploy
    ▼
Vertex-on-Edge CLI
    │
    ├── Detect framework
    ├── Select deployment configuration
    ├── Generate Dockerfile
    ├── Generate GitHub Actions
    ├── Generate Terraform
    ├── Configure GitHub
    └── Deploy infrastructure
              │
              ▼
          AWS Account
              │
       ┌──────┴──────┐
       ▼             ▼
     ECR            EC2
       │             │
       └──────┬──────┘
              ▼
        Running App
After the initial setup:
git push
   │
   ▼
GitHub Actions
   │
   ├── Checkout
   ├── Build
   ├── Test
   ├── Build Docker Image
   ├── Push Image
   └── Deploy
          │
          ▼
       AWS EC2
          │
          ▼
     Live Application
Supported Project Detection
Vertex-on-Edge detects projects by inspecting files in the selected project directory.
File
Detected Technology
package.json
Node.js
next.config.*
Next.js
vite.config.*
Vite
requirements.txt
Python
pyproject.toml
Python
go.mod
Go
pom.xml
Java / Maven
build.gradle
Java / Gradle
The detected framework determines which predefined templates are used.
Template Engine
Vertex-on-Edge uses predefined and tested templates instead of requiring developers to manually write Dockerfiles or GitHub Actions workflows.
templates/
├── docker/
│   ├── nextjs/
│   │   └── Dockerfile
│   ├── node/
│   │   └── Dockerfile
│   ├── python/
│   │   └── Dockerfile
│   └── go/
│       └── Dockerfile
│
├── github/
│   ├── nextjs.yml
│   ├── node.yml
│   ├── python.yml
│   └── go.yml
│
└── terraform/
    └── aws/
Infrastructure
Terraform is used for cloud infrastructure provisioning.
For AWS deployments, Vertex-on-Edge can provision components such as:
AWS Account
│
├── VPC
├── Public Subnet
├── Internet Gateway
├── Route Table
├── Security Group
├── EC2 Instance
├── IAM Role
├── IAM Instance Profile
├── GitHub Actions OIDC Provider
├── GitHub Actions IAM Role
└── Amazon ECR Repository
Terraform can expose deployment information such as:
application_url
instance_public_ip
instance_public_dns
registry_url
github_actions_role_arn
GitHub Actions + AWS OIDC
Vertex-on-Edge uses GitHub Actions OIDC for AWS authentication.
Instead of storing long-lived AWS access keys inside GitHub, the workflow requests a temporary AWS identity through GitHub's OIDC token.
GitHub Actions
      │
      │ OIDC Token
      ▼
GitHub OIDC Provider
      │
      ▼
AWS STS
      │
      │ AssumeRoleWithWebIdentity
      ▼
Deployment IAM Role
      │
      ├── ECR
      ├── EC2
      └── Deployment Resources
The IAM trust policy is restricted to the intended GitHub repository.
CLI
Start the CLI:
vertex-on-edge
Deploy an existing project:
vertex-on-edge deploy
The CLI guides the user through:
Project
   ↓
Framework
   ↓
Testing
   ↓
Cloud Provider
   ↓
Deployment Target
   ↓
GitHub Authentication
   ↓
Cloud Authentication
   ↓
Infrastructure Configuration
   ↓
Terraform
   ↓
GitHub Actions
   ↓
Deployment
Project Structure
vertex-on-edge/
│
├── cmd/
│   └── vertex-on-edge/
│       └── main.go
│
├── internal/
│   ├── detector/
│   ├── picker/
│   ├── templates/
│   ├── github/
│   ├── terraform/
│   ├── aws/
│   ├── docker/
│   ├── pipeline/
│   └── secrets/
│
├── templates/
│   ├── docker/
│   ├── github/
│   └── terraform/
│
├── pkg/
├── go.mod
├── go.sum
├── README.md
└── LICENSE
Cloud Providers
Current Target
AWS
└── EC2
Planned
AWS
├── EC2
└── EKS

Google Cloud
└── Cloud Run

Microsoft Azure
└── Container Apps
Security
Vertex-on-Edge is designed to avoid putting cloud credentials into source code.
Security principles:
Do not commit cloud credentials
Prefer OIDC over long-lived AWS credentials
Use GitHub encrypted Actions secrets where required
Restrict IAM trust policies to the target repository
Keep Terraform state protected
Never print secrets in CLI logs
Never write private keys into Git
Validate cloud configuration before deployment
Use least-privilege IAM policies where possible
Example
From an existing project:
cd my-next-app
vertex-on-edge deploy
Example output:
✓ Project detected
✓ Next.js detected
✓ Docker template selected
✓ Dockerfile generated
✓ GitHub Actions workflow generated
✓ Terraform configuration prepared
✓ AWS configuration validated
✓ GitHub configuration validated
✓ Infrastructure provisioned
✓ ECR repository ready
✓ EC2 instance ready
✓ OIDC configuration ready
✓ GitHub Actions configured
Development
Clone the repository:
git clone https://github.com/yashchokhat/vertex-on-edge-cli.git
cd vertex-on-edge-cli
Install dependencies:
go mod download
Build:
go build -o vertex-on-edge .
Run locally:
./vertex-on-edge
Or:
go run .
Requirements
For development:
Go
Git
Docker
Terraform
AWS CLI
GitHub account
AWS account
For AWS deployment:
AWS credentials
GitHub authentication
AWS IAM permissions
Terraform
Docker
Roadmap
Phase 1 — Core CLI
Go CLI
Interactive project selection
Framework detection
Template engine
Dockerfile generation
Local build validation
Test integration
Phase 2 — GitHub Integration
GitHub authentication
Repository detection
Repository creation
GitHub Actions generation
GitHub Actions secrets
OIDC configuration
Automated repository synchronization
Phase 3 — Cloud Deployment
Terraform engine
AWS account validation
ECR provisioning
EC2 provisioning
IAM configuration
Automated application deployment
Health-check validation
Phase 4 — Multi-Cloud
Google Cloud
Azure
Additional container platforms
Advanced deployment strategies
Phase 5 — Advanced Features
Automatic rollback
Post-deployment health checks
Cost estimation
Deployment status
Optional web dashboard
Kubernetes support
Design Principle
Vertex-on-Edge is a bootstrapper, not a runtime dependency.
The CLI performs the initial setup and then gets out of the way.
The developer owns:
Source Code
     +
GitHub Repository
     +
CI/CD Pipeline
     +
Cloud Account
     +
Infrastructure
Vertex-on-Edge simply automates the configuration between them.
License
This project is open source. See LICENSE for details.