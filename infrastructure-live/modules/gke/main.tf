data "google_client_config" "default" {} # Get confg with project, access token, region

provider "kubernetes" {
  host                   = "https://${module.gke.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(module.gke.ca_certificate)
}

module "gke" {
  source                     = "terraform-google-modules/kubernetes-engine/google"
  project_id                 = var.project
  name                       = "gke-cluster-1"
  region                     = var.region
  zones                      = ["${var.region}-a", "${var.region}-b", "${var.region}-c"]
  network                    = var.network
  subnetwork                 = var.subnetwork
  ip_range_pods              = "pods-range"
  ip_range_services          = "svc-range"
  http_load_balancing        = false
  network_policy             = false
  horizontal_pod_autoscaling = true
  filestore_csi_driver       = false
  dns_cache                  = false

  node_pools = [
    {
      name                        = "triple-pool" 
      machine_type                = "e2-medium"
      # one zone 
      min_count                   = 1
      max_count                   = 6
      local_ssd_count             = 0
      spot                        = false
      disk_size_gb                = 20
      disk_type                   = "pd-standard"
      image_type                  = "COS_CONTAINERD"
      enable_gcfs                 = false
      enable_gvnic                = false
      logging_variant             = "DEFAULT"
      auto_repair                 = true
      auto_upgrade                = true
      preemptible                 = false
      initial_node_count          = 1
      deletion_protection         = false   
    }
  ]

# Rules for SA 
  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

# Labels on node pools, can be used for scheduling pods to specific node pools
  node_pools_labels = {
    all = {}

    standard-pool = {
      role = "standard"
      type = "development"
    }
  }

# Metadata (keyword for instances in pool)
  node_pools_metadata = {
    all = {}
    main-pool = {
      node-pool-metadata-custom-value = "my-node-pool"
    }
  }

  node_pools_tags = {
    all = []

    main-pool = [
      "default-node-pool",
    ]
  }
}