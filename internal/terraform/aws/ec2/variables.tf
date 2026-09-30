variable "aws_region" {
  description = "The AWS region to deploy to"
  type        = string
}

variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "app_port" {
  description = "Port the application runs on"
  type        = number
  default     = 3000
}

variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
}

variable "github_repository" {
  description = "GitHub repository name"
  type        = string
}

variable "create_oidc_provider" {
  description = "Whether to create the GitHub OIDC provider (set to false if it already exists in this AWS account)"
  type        = bool
  default     = true
}

variable "oidc_provider_arn" {
  description = "ARN of an existing GitHub OIDC provider (only used when create_oidc_provider is false)"
  type        = string
  default     = ""
}

variable "create_vpc" {
  description = "Whether to create a new VPC or use the default VPC"
  type        = bool
  default     = true
}

variable "create_subnet" {
  description = "Whether to create a new subnet or use a default subnet"
  type        = bool
  default     = true
}

variable "existing_key_pair_name" {
  description = "Name of an existing AWS key pair to use for SSH access (leave empty to auto-generate one)"
  type        = string
  default     = ""
}
