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

variable "cluster_name" {
  description = "The name of the cluster"
  type        = string
}

variable "cloudflare_auth_domain" {
  description = "The CloudFlare Access authentication domain (e.g. https://example.cloudflareaccess.com)"
  type        = string
}

variable "cloudflare_domain" {
  description = "The CloudFlare Access domain (e.g. trident.operator.dev)"
  type        = string
}

variable "cloudflare_audience" {
  description = "The CloudFlare Access audience (64 hex chars)"
  type        = string
}

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------

variable "worker_image" {
  description = "The container image ref for worker"
  type        = string
}

variable "dispatcher_image" {
  description = "The container image ref for dispatcher"
  type        = string
}

variable "orchestrator_image" {
  description = "The container image ref for orchestrator"
  type        = string
}
