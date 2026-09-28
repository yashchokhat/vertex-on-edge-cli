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
  description = "Whether to create the GitHub OIDC provider"
  type        = bool
  default     = true
}

variable "oidc_provider_arn" {
  description = "ARN of the existing GitHub OIDC provider"
  type        = string
  default     = ""
}
