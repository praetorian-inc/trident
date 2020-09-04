# -----------------------------------------------------------------------------
# REQUIRED PARAMETERS
# These parameters must be supplied when consuming this module.
# -----------------------------------------------------------------------------

variable "project" {
  description = "The GCP project ID"
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

variable "pubsub_topic" {
  description = "The name of the PubSub topic to publish results to"
  type        = string
}

variable "pubsub_subscription" {
  description = "The name of the PubSub subscription to fetch jobs from"
  type        = string
}

variable "worker_url" {
  description = "The URL to a webhook worker for job submission"
  type        = string
}

variable "worker_token" {
  description = "The access token to use when authenticating to a worker"
  type        = string
}

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------

variable "worker_name" {
  description = "The type of worker to submit jobs to"
  type        = string
  default     = "webhook"
}
