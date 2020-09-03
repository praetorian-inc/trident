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

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------
