resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "instance" {
  project = "terraform-test-288321"
  name = "${var.database_instance_name}-${random_id.db_name_suffix.hex}"
  database_version = var.database_version
  region = var.database_region

  settings {
    tier = var.database_machine_type

    ip_configuration {
      ipv4_enabled = false
      private_network = var.private_network_id
    }
  }
}
