output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.app_server.id
}

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

output "ssh_private_key" {
  description = "Private key for SSH access (only available when auto-generating a key pair)"
  value       = var.existing_key_pair_name == "" ? tls_private_key.ssh_key[0].private_key_pem : "Using existing key pair: ${var.existing_key_pair_name}"
  sensitive   = true
}
