
locals {
  name = "dispatcher"
 
  labels = {
    app = "dispatcher"
  }
 
  env = {
    "PROJECT_ID"      = var.project,
    "RESULT_TOPIC_ID" = var.pubsub_topic,
    "SUBSCRIPTION_ID" = var.pubsub_subscription,
    "WORKER_NAME"     = var.worker_name,
  }

  worker_config = jsonencode({
    "url"   = var.worker_url,
    "token" = var.worker_token,
  })
}

resource "google_service_account" "server" {
  project      = var.project
  account_id   = "trident-${local.name}-sa"
  display_name = "Trident ${local.name}"
}

resource "google_service_account_iam_member" "workload-identity" {
  service_account_id = google_service_account.server.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project}.svc.id.goog[${var.namespace}/${local.name}]"
}

resource "google_pubsub_topic_iam_member" "server-topic" {
  topic  = var.pubsub_topic
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${google_service_account.server.email}"
}

resource "google_pubsub_subscription_iam_member" "server-subscription" {
  subscription = var.pubsub_subscription
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${google_service_account.server.email}"
}

resource "kubernetes_service_account" "server" {
  metadata {
    namespace   = var.namespace
    name        = local.name
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.server.email
    }
  }
}

resource "kubernetes_config_map" "worker" {
  metadata {
    namespace = var.namespace
    name      = "dispatcher-worker"
  }

  data = {
    config = local.worker_config
  }
}

resource "kubernetes_deployment" "server" {
  metadata {
    namespace = var.namespace
    name      = local.name
    labels    = local.labels
  }

  spec {
    replicas = 20

    selector {
      match_labels = local.labels
    }

    template {
      metadata {
        labels = local.labels
      }

      spec {
        service_account_name = kubernetes_service_account.server.metadata[0].name
        container {
          image = "${var.image}:${var.image_version}"
          name  = local.name

          dynamic "env" {
            for_each = local.env
            content {
              name  = env.key
              value = env.value
            }
          }

          env {
            name = "WORKER_CONFIG"
            value_from {
              config_map_key_ref {
                name = kubernetes_config_map.worker.metadata[0].name
                key  = "config"
              }
            }
          }
        }
      }
    }
  }
}
