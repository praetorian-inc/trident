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

# -----------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# -----------------------------------------------------------------------------

variable "database_instance_name" {
  default = "backend-databases"
}

variable "database_version" {
  default = "POSTGRES_12"
}

variable "database_machine_type" {
  default = "db-g1-small"
}

//this default will cause it to fail
variable "private_network_id" {
  default = "test"
}
