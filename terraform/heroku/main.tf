resource "random_integer" "random" {
  min = 1
  max = 9999999
}

resource "heroku_app" "app" {
  name   = "db1000n-${random_integer.random.result}"
  region = var.region

  config_vars = {
    GOVERSION = "1.17"
  }

  buildpacks = [
    "heroku/go"
  ]
}

resource "heroku_build" "build" {
  app_id = heroku_app.app.id

  source {
    url = "${var.repo}/archive/v${var.app_version}.tar.gz"
  }
}

resource "heroku_formation" "formation" {
  app_id     = heroku_app.app.id
  type       = "worker"
  quantity   = var.instance_count
  size       = var.instance_type
  depends_on = [heroku_build.build]
}
