
resource "random_id" "token" {
  byte_length = 16
}

resource "google_cloud_run_service" "worker" {
  project  = var.project
  location = var.location
  name     = "webhook-worker"

  template {
    spec {
      containers {
        image = var.image

        env {
          name = "ACCESS_TOKEN"
          value = random_id.token.hex
        }
      }
    }
  }

  autogenerate_revision_name = true
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  project  = google_cloud_run_service.worker.project
  location = google_cloud_run_service.worker.location
  service  = google_cloud_run_service.worker.name

  policy_data = data.google_iam_policy.noauth.policy_data
}
