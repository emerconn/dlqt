resource "azurerm_resource_group" "this" {
  name     = "rg-dlqt"
  location = "centralus"
}

# ===== Service Bus =====

resource "azurerm_servicebus_namespace" "this" {
  name                = "sb-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "Basic"
}

resource "azurerm_servicebus_queue" "this" {
  name         = "sbq-dlqt-1"
  namespace_id = azurerm_servicebus_namespace.this.id
}

# ===== Container App (API) =====

resource "azurerm_log_analytics_workspace" "this" {
  name                = "law-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_container_app_environment" "this" {
  name                = "cae-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name

  logs_destination           = "log-analytics"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
}

resource "azurerm_container_app" "api" {
  name                         = "ca-dlqt-api"
  container_app_environment_id = azurerm_container_app_environment.this.id
  resource_group_name          = azurerm_resource_group.this.name
  revision_mode                = "Single"

  template {
    container {
      name   = "api"
      image  = "ghcr.io/emerconn/dlqt/api:latest"
      cpu    = 0.25
      memory = "0.5Gi"
    }
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  identity {
    type = "SystemAssigned"
  }

  lifecycle {
    ignore_changes = [
      template[0].container[0].image,
    ]
  }
}

resource "azurerm_role_assignment" "api_app_reg_service_bus_data_owner" {
  scope                = azurerm_servicebus_namespace.this.id
  role_definition_name = "Azure Service Bus Data Owner"
  principal_id         = azurerm_container_app.api.identity[0].principal_id
}

resource "azurerm_role_assignment" "me_service_bus_data_owner" {
  scope                = azurerm_servicebus_namespace.this.id
  role_definition_name = "Azure Service Bus Data Owner"
  principal_id         = local.my_principal_id
}
