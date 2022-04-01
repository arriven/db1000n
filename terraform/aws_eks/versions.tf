terraform {
  required_version = ">= 1.0.9"
  required_providers {
    aws = {
      version = "= 4.5.0"
    }
    kubernetes = {
      version = ">= 2.9.0"
    }
  }
}
