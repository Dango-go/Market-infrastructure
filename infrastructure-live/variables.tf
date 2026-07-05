variable "github_repo" {
    type = string
    default = "Dango-go/Market-infrastructure" 
    description = "GitHub repository name in the format 'owner/repo' for which the service account will be used."
}
variable "project" {
    type = string
    default = "market-infrastructure"
}
    
variable "region" {
    type = string
    default = "europe-central2"
}

variable "network_name" {
    type = string
    description = "The name of the VPC network"
    default = "main-vpc"
}

variable "subnetwork_name" {
    type = string
    description = "The name of the subnetwork"
    default = "subnet-1"
}

variable "ip_cidr_range" {
    type = string
    description = "The IP CIDR range for the subnetwork"
    default = "10.10.1.0/24"
}

variable "principalset" {
    type = string
    default = "principalSet://iam.googleapis.com/projects/116938195044/locations/global/workloadIdentityPools/github-pool-v1/attribute.repository/Dango-go/Market-infrastructure"
}
