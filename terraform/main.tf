
terraform {
  required_version = ">= 0.13"
}

provider "google" {
  version = "~> 3.37.0"

  project = var.project
  region  = var.location
}

provider "google-beta" {
  version = "~> 3.37.0"

  project = var.project
  region  = var.location
}

data "google_client_config" "provider" {}

data "google_project" "project" {
  project_id = var.project
}

provider "random" {}

resource "google_project_service" "services" {
  project = var.project
  for_each = toset([
    "container.googleapis.com",
    "run.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false
}

module "gke_cluster" {
  # When using these modules in your own templates, you will need to use a Git
  # URL with a ref attribute that pins you to a specific version of the modules,
  # such as the following example:
  # source = "github.com/praetorian-inc/trident.git//terraform/modules/gke-cluster?ref=v0.1.0"
  source = "./modules/gke-cluster"

  project  = var.project
  location = var.location
  name     = var.cluster_name

  depends_on = [
    google_project_service.services["container.googleapis.com"],
  ]
}

provider "kubernetes" {
  version = "~> 1.13.0"

  load_config_file = false

  host  = "https://${module.gke_cluster.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = module.gke_cluster.cluster_ca_certificate
}

resource "kubernetes_namespace" "ns" {
  metadata {
    name = "trident"
  }
}

module "pubsub" {
  source = "./modules/pubsub"

  project = var.project
}

module "worker" {
  source = "./modules/worker"

  project  = var.project
  location = var.location

  depends_on = [
    google_project_service.services["run.googleapis.com"],
  ]
}

module "dispatcher" {
  source = "./modules/dispatcher"

  project = var.project

  pubsub_topic        = module.pubsub.pubsub_topic_results
  pubsub_subscription = module.pubsub.pubsub_subscription_credentials

  worker_url   = module.worker.endpoint
  worker_token = module.worker.token

  depends_on = [
    google_project_service.services["container.googleapis.com"],
  ]
}
