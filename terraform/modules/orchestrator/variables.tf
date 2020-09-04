# -----------------------------------------------------------------------------
# REQUIRED PARAMETERS
# These parameters must be supplied when consuming this module.
# -----------------------------------------------------------------------------

variable "project" {
  description = "The GCP project ID"
  type        = string
}

variable "location" {
  description = "The location (region or zone) to host the cluster in"
  type        = string
}

variable "namespace" {
  description = "The Kubernetes namespace for deployment"
  type        = string
}

variable "image" {
  description = "The container image to deploy"
  type        = string
}

variable "network" {
  description = "The VPC network ID for redis"
  type        = string
}

variable "pubsub_topic" {
  description = "The name of the PubSub topic to publish credentials to"
  type        = string
}

variable "pubsub_subscription" {
  description = "The name of the PubSub subscription to fetch results from"
  type        = string
}

variable "db_instance_name" {
  description = "The name of the Cloud SQL database instance"
  type        = string
}

variable "db_connection_name" {
  description = "The connection name of the Cloud SQL database instance"
  type        = string
}

variable "cloudflare_auth_domain" {
  description = "The CloudFlare access authentication domain"
  type        = string
}

variable "cloudflare_domain" {
  description = "The CloudFlare domain to host this service on"
  type        = string
}

variable "cloudflare_audience" {
  description = "The CloudFlare access audience"
  type        = string
}

variable "cloudflare_cert" {
  description = "The CloudFlare certificate and private key blob"
  type        = string
}

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------
