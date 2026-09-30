provider "aws" {
  region = var.aws_region
}

resource "random_id" "suffix" {
  byte_length = 3
}

# ---------------------------------------------------------------------------
# Networking: VPC, Subnet, Internet Gateway, Route Table
# ---------------------------------------------------------------------------

# Use the default VPC when create_vpc is false.
data "aws_vpc" "default" {
  count   = var.create_vpc ? 0 : 1
  default = true
}

resource "aws_vpc" "main" {
  count                = var.create_vpc ? 1 : 0
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name      = "${var.project_name}-vpc"
    ManagedBy = "Vertex-on-Edge"
  }
}

locals {
  vpc_id = var.create_vpc ? aws_vpc.main[0].id : data.aws_vpc.default[0].id
}

# Internet Gateway (only needed for a new VPC; the default VPC already has one).
resource "aws_internet_gateway" "main" {
  count  = var.create_vpc ? 1 : 0
  vpc_id = aws_vpc.main[0].id

  tags = {
    Name      = "${var.project_name}-igw"
    ManagedBy = "Vertex-on-Edge"
  }
}

# Use an existing subnet when create_subnet is false.
data "aws_subnets" "default" {
  count = var.create_subnet ? 0 : 1
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
  filter {
    name   = "default-for-az"
    values = ["true"]
  }
}

resource "aws_subnet" "public" {
  count                   = var.create_subnet ? 1 : 0
  vpc_id                  = local.vpc_id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name      = "${var.project_name}-subnet-public"
    ManagedBy = "Vertex-on-Edge"
  }
}

locals {
  subnet_id = var.create_subnet ? aws_subnet.public[0].id : data.aws_subnets.default[0].ids[0]
}

# Route table for the new subnet (only when creating a new VPC).
resource "aws_route_table" "public" {
  count  = var.create_vpc && var.create_subnet ? 1 : 0
  vpc_id = local.vpc_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main[0].id
  }

  tags = {
    Name      = "${var.project_name}-rt"
    ManagedBy = "Vertex-on-Edge"
  }
}

resource "aws_route_table_association" "public" {
  count          = var.create_vpc && var.create_subnet ? 1 : 0
  subnet_id      = local.subnet_id
  route_table_id = aws_route_table.public[0].id
}

# ---------------------------------------------------------------------------
# Security Group
# ---------------------------------------------------------------------------

resource "aws_security_group" "app_sg" {
  name        = "${var.project_name}-sg-${random_id.suffix.hex}"
  description = "Allow inbound traffic for application"
  vpc_id      = local.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = var.app_port
    to_port     = var.app_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name      = "${var.project_name}-sg"
    ManagedBy = "Vertex-on-Edge"
  }
}

# ---------------------------------------------------------------------------
# ECR Repository
# ---------------------------------------------------------------------------

resource "aws_ecr_repository" "app" {
  name                 = "${var.project_name}-${random_id.suffix.hex}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    ManagedBy   = "Vertex-on-Edge"
    Project     = var.project_name
    Environment = "production"
  }
}

# ---------------------------------------------------------------------------
# EC2 IAM Role (allows the instance to pull from ECR and receive SSM commands)
# ---------------------------------------------------------------------------

resource "aws_iam_role" "ec2_role" {
  name = "${var.project_name}-ec2-role-${random_id.suffix.hex}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    ManagedBy = "Vertex-on-Edge"
  }
}

resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "ssm_managed" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "ec2_profile" {
  name = "${var.project_name}-profile-${random_id.suffix.hex}"
  role = aws_iam_role.ec2_role.name
}

# ---------------------------------------------------------------------------
# SSH Key Pair
# ---------------------------------------------------------------------------

# Auto-generate a key pair when the user does not supply an existing key pair name.
resource "tls_private_key" "ssh_key" {
  count     = var.existing_key_pair_name == "" ? 1 : 0
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "app_key" {
  count      = var.existing_key_pair_name == "" ? 1 : 0
  key_name   = "${var.project_name}-key-${random_id.suffix.hex}"
  public_key = tls_private_key.ssh_key[0].public_key_openssh
}

locals {
  key_pair_name = var.existing_key_pair_name != "" ? var.existing_key_pair_name : aws_key_pair.app_key[0].key_name
}

# ---------------------------------------------------------------------------
# EC2 Instance
# ---------------------------------------------------------------------------

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

resource "aws_instance" "app_server" {
  ami           = data.aws_ami.amazon_linux_2023.id
  instance_type = var.instance_type
  subnet_id     = local.subnet_id
  key_name      = local.key_pair_name

  vpc_security_group_ids = [aws_security_group.app_sg.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2_profile.name

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y docker amazon-ssm-agent
              systemctl start docker
              systemctl enable docker
              systemctl start amazon-ssm-agent
              systemctl enable amazon-ssm-agent
              usermod -aG docker ec2-user
              EOF

  tags = {
    Name      = "${var.project_name}-server"
    ManagedBy = "Vertex-on-Edge"
  }
}

# ---------------------------------------------------------------------------
# GitHub Actions OIDC Provider (global singleton per AWS account)
# ---------------------------------------------------------------------------

resource "aws_iam_openid_connect_provider" "github" {
  count           = var.create_oidc_provider ? 1 : 0
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd",
    "1b511abead59c6ce207077c0bf0e0043b1382612",
    "06d927fecd0a84aeba28aad1d808139470fe95c3",
    "ffffffffffffffffffffffffffffffffffffffff",
  ]
}

# ---------------------------------------------------------------------------
# GitHub Actions IAM Role (this is the role GitHub Actions assumes via OIDC)
#
# This is the fix for:
#   "Error: Could not assume role with OIDC: Not authorized to perform
#    sts:AssumeRoleWithWebIdentity"
#
# The trust policy below explicitly allows the OIDC provider to issue
# temporary credentials, scoped to only this repository and branch.
# ---------------------------------------------------------------------------

data "aws_caller_identity" "current" {}

locals {
  oidc_provider_arn = var.create_oidc_provider ? aws_iam_openid_connect_provider.github[0].arn : var.oidc_provider_arn
}

resource "aws_iam_role" "github_actions_role" {
  name = "${var.project_name}-gh-actions-${random_id.suffix.hex}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = local.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_owner}*"
          }
        }
      }
    ]
  })

  tags = {
    ManagedBy = "Vertex-on-Edge"
    Project   = var.project_name
  }
}

# The GitHub Actions role needs permission to push images to ECR,
# send SSM commands to EC2, and read EC2 metadata.

resource "aws_iam_role_policy" "github_actions_ecr_push" {
  name = "${var.project_name}-ecr-push"
  role = aws_iam_role.github_actions_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
        ]
        Resource = aws_ecr_repository.app.arn
      },
    ]
  })
}

resource "aws_iam_role_policy" "github_actions_ssm" {
  name = "${var.project_name}-ssm-deploy"
  role = aws_iam_role.github_actions_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:SendCommand",
          "ssm:GetCommandInvocation",
        ]
        Resource = [
          "arn:aws:ec2:${var.aws_region}:${data.aws_caller_identity.current.account_id}:instance/${aws_instance.app_server.id}",
          "arn:aws:ssm:${var.aws_region}::document/AWS-RunShellScript",
        ]
      }
    ]
  })
}
