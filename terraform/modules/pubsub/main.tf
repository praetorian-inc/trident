
resource "google_pubsub_topic" "credentials" {
  project = var.project
  name    = "trident-credentials"
}

resource "google_pubsub_topic" "results" {
  project = var.project
  name    = "trident-results"
}

resource "google_pubsub_subscription" "credentials" {
  project = var.project
  name    = "trident-dispatcher-credentials"
  topic   = google_pubsub_topic.credentials.id

  ack_deadline_seconds = 300
}

resource "google_pubsub_subscription" "results" {
  project = var.project
  name    = "trident-orchestrator-results"
  topic   = google_pubsub_topic.results.id

  ack_deadline_seconds = 300
}
