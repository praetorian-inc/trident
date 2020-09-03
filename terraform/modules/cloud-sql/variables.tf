variable "database_instance_name" {
  default = "backend-databases"
}

variable "database_version" {
  default = "POSTGRES_12"
}

variable "database_region" {
  default = "us-central1"
}

variable "database_machine_type" {
  default = "db-g1-small"
}

//this default will cause it to fail
variable "private_network_id" {
  default = "test"
}
