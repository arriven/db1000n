resource "awslightsail_container_service" "service" {
  name  = "${var.app}-service"
  power = var.power
  scale = var.scale
}

resource "awslightsail_container_deployment" "deployment" {
  container_service_name = awslightsail_container_service.service.id
  container {
    container_name = "${var.app}-deployment"
    image          = var.image

    environment {
      key   = "ENABLE_PRIMITIVE"
      value = "false"
    }
  }
}
