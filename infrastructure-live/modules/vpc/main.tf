resource "google_compute_network" "main" {
  name                    = var.network_name
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"
}

resource "google_compute_subnetwork" "subnet" {
  name          = var.subnetwork_name
  ip_cidr_range = var.cidr_range  # Nodes
  region        = var.region
  network       = google_compute_network.main.self_link
  secondary_ip_range {
    range_name    = "pods-range"
    ip_cidr_range = "10.20.0.0/16"            # Pods
  }
  secondary_ip_range {
    range_name    = "svc-range"
    ip_cidr_range = "10.50.0.0/16"            # Services
  }
}

resource "google_compute_firewall" "standard-firewall" {
  name    = "allow-ssh"
  network = google_compute_network.main.name  # Firewall attached to the VPC network

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["firewall-label"]   
}