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

resource "google_cloud_run_v2_service" "my_service" {
  name = "oidc-server"
  location = "asia-northeast1"
  deletion_protection = false
  ingress = "INGRESS_TRAFFIC_ALL"

  scaling {
    max_instance_count = 2
  }

  template {
    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
    }
  }
}
