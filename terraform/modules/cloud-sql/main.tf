
resource "random_id" "db_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "instance" {
  project = var.project
  region  = var.location
  name    = "${var.database_instance_name}-${random_id.db_suffix.hex}"

  database_version = var.database_version

  settings {
    tier = var.database_machine_type

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.private_network_id
    }
  }
}
