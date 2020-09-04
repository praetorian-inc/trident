
output "pubsub_topic_credentials" {
  description = "The name of credentials topic"
  value = google_pubsub_topic.credentials.name
}

output "pubsub_topic_results" {
  description = "The name of results topic"
  value = google_pubsub_topic.results.name
}

output "pubsub_subscription_credentials" {
  description = "The name of credentials subscription (used by dispatcher)"
  value = google_pubsub_subscription.credentials.name
}

output "pubsub_subscription_results" {
  description = "The name of results subscription (used by orchestrator consumer)"
  value = google_pubsub_subscription.results.name
}
