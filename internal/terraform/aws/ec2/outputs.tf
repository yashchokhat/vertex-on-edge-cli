output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.app_server.public_ip
}

output "instance_public_dns" {
  description = "Public DNS of the EC2 instance"
  value       = aws_instance.app_server.public_dns
}

output "application_url" {
  description = "URL to access the application"
  value       = "http://${aws_instance.app_server.public_dns}:${var.app_port}"
}

output "registry_url" {
  description = "ECR Repository URL"
  value       = aws_ecr_repository.app.repository_url
}

output "ssh_private_key" {
  description = "Private key for SSH access (sensitive)"
  value       = tls_private_key.ssh_key.private_key_pem
  sensitive   = true
}

