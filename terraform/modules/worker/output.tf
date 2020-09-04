
output "endpoint" {
  value = google_cloud_run_service.worker.status[0].url
}

output "token" {
  value = random_id.token.hex
}
