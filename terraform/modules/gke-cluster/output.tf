
output "name" {
  description = "The name of the cluster master. This output is used for interpolation with node pools, other modules."
  value       = google_container_cluster.cluster.name
}

output "node_pool" {
  description = "The id of the node pool. This output is used for interpolation with other modules."
  value       = google_container_node_pool.default.id
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
