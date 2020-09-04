
output "config" {
  description = "The URI of the deployed orchestrator instance"
  value       = jsonencode({
    "orchestrator-url" = "https://${var.cloudflare_domain}"
  })
}
