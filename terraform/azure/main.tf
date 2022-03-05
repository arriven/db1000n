provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "${var.prefix}-rg"
  location = "northeurope"
}

module "bomblet" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "northeurope"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}

module "bomblet_we" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "westeurope"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}

module "bomblet_cc" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "canadacentral"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}

module "bomblet_uae" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "uaenorth"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}

module "bomblet_cu" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "centralus"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}

module "bomblet_ea" {
  source = "./bomblet"

  bomblet_count       = var.bomblet_count
  region              = "eastasia"
  prefix              = var.prefix
  resource_group_name = azurerm_resource_group.main.name
}