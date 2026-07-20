module "gke" {
    source = "./modules/gke"
    region = var.region
    project = var.project
    network = var.network_name
    subnetwork = module.vpc.subnet_name
    service_account_email = "default" 
}

module "vpc" {
    source = "./modules/vpc"
    region = var.region
    network_name = var.network_name
    subnetwork_name = var.subnetwork_name
    cidr_range = var.ip_cidr_range
}

module "iam-roles" {
    source = "./modules/iam-roles"
    github_repo = var.github_repo
    project_name = var.project
    principal = var.principalset
}