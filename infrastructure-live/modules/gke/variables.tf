variable "region" {
    type = string
}

variable "project" {
    type = string
    default = "infrastructure-open-telemetry"
}

variable "network" {
    type = string
}

variable "subnetwork" {
    type = string
}

variable "service_account_email" {
    type = string
    description = "The email of the service account to be used by the GKE nodes."
}