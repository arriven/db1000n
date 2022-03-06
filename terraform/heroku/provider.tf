terraform {
  required_providers {
    heroku = {
      source  = "heroku/heroku"
      version = "~> 5.0"
    }
  }
  required_version = "~> 1.0"
}

provider "heroku" {
  email   = var.email
  api_key = var.api_key
}
