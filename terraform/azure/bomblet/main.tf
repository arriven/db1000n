resource "azurerm_container_group" "main" {
  count               = var.bomblet_count
  name                = format("%s-%s", "${var.prefix}-${var.region}", format("%02d", count.index + 1))
  location            = var.region
  resource_group_name = var.resource_group_name
  ip_address_type     = "None"
  os_type             = "Linux"
  restart_policy      = "Always"

  container {
    name   = "main"
    image  = var.attack_image
    cpu    = var.attack_cpu
    memory = var.attack_memory

    environment_variables = var.attack_environment_variables
  }
}

