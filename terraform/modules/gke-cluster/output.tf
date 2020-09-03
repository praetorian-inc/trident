
output "name" {
  description = "The name of the cluster master. This output is used for interpolation with node pools, other modules."
  value       = google_container_cluster.cluster.name
}

output "endpoint" {
  description = "The IP address of the cluster master."
  sensitive   = true
  value       = google_container_cluster.cluster.endpoint
}

output "cluster_ca_certificate" {
  description = "The public certificate that is the root of trust for the cluster."
  value       = base64decode(google_container_cluster.cluster.master_auth[0].cluster_ca_certificate)
}
