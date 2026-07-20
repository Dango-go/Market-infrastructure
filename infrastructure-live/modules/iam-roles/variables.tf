variable "github_repo" {
    description = "GitHub repository name in the format 'owner/repo'"
    type        = string
}

variable "project_name" {
    description = "GitHub repository name in the format 'owner/repo'"
    type        = string
}

variable "principal" {
    description = "The principal (service account or workload identity) that will be granted the IAM role"
    type        = string
}