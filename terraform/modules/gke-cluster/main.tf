
resource "google_container_cluster" "cluster" {
  provider = google-beta

  project  = var.project
  location = var.location
  name     = var.name

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1

  networking_mode = "VPC_NATIVE"
  network    = var.network
  subnetwork = var.subnetwork

  ip_allocation_policy {
    cluster_ipv4_cidr_block  = "/16"
    services_ipv4_cidr_block = "/22"
  }

  release_channel {
    channel = var.release_channel
  }

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }

  enable_shielded_nodes = true

  workload_identity_config {
    identity_namespace = "${var.project}.svc.id.goog"
  }

}

resource "google_container_node_pool" "default" {
  project  = var.project
  location = var.location

  cluster = google_container_cluster.cluster.name
  name    = "default"

  node_count = 1

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]

    metadata = {
      disable-legacy-endpoints = "true"
    }
  }

}
