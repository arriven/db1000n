terraform {
  required_providers {
    awslightsail = {
      source  = "DeYoungTech/awslightsail"
      version = "0.7.0"
    }
  }
  required_version = "~> 1.0"
}

provider "awslightsail" {
  region = var.region_name
}
