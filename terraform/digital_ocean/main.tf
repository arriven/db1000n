resource "digitalocean_app" "db1000n" {
  spec {
    name   = "db1000n"
    region = var.region

    worker {
      name               = "db1000n-service"
      environment_slug   = "go"
      instance_count     = var.instance_count
      instance_size_slug = var.instance_size_slug

      git {
        repo_clone_url = var.repo
        branch         = "main"
      }
    }
  }
}
