# For Runner SA to push images to Artifact Registry GCP
resource "google_service_account" "github_sa" {
  account_id   = "github-ci"
  display_name = "github-ci"
}

resource "google_iam_workload_identity_pool" "pool" {
  workload_identity_pool_id = "github-pool-v1"
}

resource "google_iam_workload_identity_pool_provider" "provider" {
  workload_identity_pool_id    = google_iam_workload_identity_pool.pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
  attribute_mapping = {
    "google.subject" = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }
  attribute_condition = "assertion.repository == '${var.github_repo}'"
}

resource "google_service_account_iam_member" "identity_user" {
  service_account_id = google_service_account.github_sa.name  # sa id
  role               = "roles/iam.workloadIdentityUser" # role for assuming the service account   
  member             = var.principal  # member (role + sa)
}

resource "google_project_iam_member" "artifact_registry_writer" {
  project = var.project_name
  role   = "roles/artifactregistry.writer"
  member  =  var.principal
}

resource "google_project_iam_member" "k8s_developer" {
    project = var.project_name
    role    = "roles/container.developer"
    member  = "serviceAccount:github-ci@market-infrastructure.iam.gserviceaccount.com"
}