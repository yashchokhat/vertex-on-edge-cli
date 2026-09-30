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

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y docker
              systemctl start docker
              systemctl enable docker
              usermod -aG docker ec2-user
              EOF

  tags = {
    Name      = "${var.project_name}-server"
    ManagedBy = "Vertex-on-Edge"
  }
}
