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

variable "name" {
  description = "The name of the cluster"
  type        = string
}

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------

variable "network" {
  description = "A reference (self link) to the VPC network to host the cluster in"
  type        = string
  default     = "default"
}

variable "release_channel" {
  description = "The release_channel to use for automatic updates to the cluster"
  type        = string
  default     = "STABLE"
}
