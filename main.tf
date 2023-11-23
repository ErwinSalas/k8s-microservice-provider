terraform {
  required_providers {
    k8s-microservice = {
      source = "ErwinSalas/k8s-microservice"
    }
  }
}


provider "k8s-microservice" {
  config_path       = "~/.kube/config"
  project_namespace = "terraform"
}