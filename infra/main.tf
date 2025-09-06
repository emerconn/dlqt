# Generate random UUIDs for Azure AD app registration
resource "random_uuid" "oauth2_scope_id" {}

resource "random_uuid" "app_role_id" {}

resource "azuread_application" "dlqt" {
  display_name     = "dlqt"
  description      = "DLQT - Dead Letter Queue Tool for Azure Service Bus"
  sign_in_audience = "AzureADMyOrg"

  api {
    mapped_claims_enabled          = true
    requested_access_token_version = 2

    oauth2_permission_scope {
      admin_consent_description  = "Allow the application to manage dead letter queue operations on behalf of the signed-in user."
      admin_consent_display_name = "Manage DLQ Operations"
      enabled                    = true
      id                         = random_uuid.oauth2_scope_id.result
      type                       = "User"
      user_consent_description   = "Allow the application to manage dead letter queue operations on your behalf."
      user_consent_display_name  = "Manage DLQ Operations"
      value                      = "dlq.manage.user"
    }
  }

  app_role {
    allowed_member_types = ["User"]
    description          = "Can retrigger dead letter queue messages"
    display_name         = "ServiceBus DLQ Retrigger"
    enabled              = true
    id                   = random_uuid.app_role_id.result
    value                = "ServiceBus.DLQRetrigger"
  }

  # Allow public clients (like Azure CLI) to authenticate
  public_client {
    redirect_uris = [
      "http://localhost",
      "urn:ietf:wg:oauth:2.0:oob"
    ]
  }

  required_resource_access {
    resource_app_id = "00000003-0000-0000-c000-000000000000" # Microsoft Graph

    resource_access {
      id   = "e1fe6dd8-ba31-4d61-89e7-88639da4683d" # User.Read
      type = "Scope"
    }
  }
}

resource "azuread_service_principal" "dlqt" {
  client_id                    = azuread_application.dlqt.client_id
  app_role_assignment_required = true
}

resource "azurerm_resource_group" "this" {
  name     = "rg-dlqt"
  location = "centralus"
}

resource "azurerm_log_analytics_workspace" "this" {
  name                = "law-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

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

resource "azurerm_container_app_environment" "this" {
  name                = "cae-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name

  logs_destination           = "log-analytics"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
}

resource "azurerm_container_app" "auth" {
  name                         = "ca-dlqt-auth"
  container_app_environment_id = azurerm_container_app_environment.this.id
  resource_group_name          = azurerm_resource_group.this.name
  revision_mode                = "Single"

  template {
    container {
      name   = "auth"
      image  = "ghcr.io/emerconn/dlqt/auth:latest"
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

resource "azurerm_role_assignment" "this" {
  scope                = azurerm_servicebus_namespace.this.id
  role_definition_name = "Azure Service Bus Data Owner"
  principal_id         = azurerm_container_app.auth.identity[0].principal_id
}
