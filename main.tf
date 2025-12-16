terraform {
    required_providers {
       google = {
        source = "hashicorp/google"
       }
    }
}

variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

provider "google" {
  project = var.project_id
  region  = "asia-northeast1"
}

resource "google_artifact_registry_repository" "my_registry" {
    location = "asia-northeast1"
    repository_id = "registry-id-1"
    format = "DOCKER"
}
