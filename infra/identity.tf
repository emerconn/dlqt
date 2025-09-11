# ===== DLQT API app reg =====

resource "random_uuid" "dlqt_api_scope_read_id" {}

resource "random_uuid" "dlqt_api_scope_retrigger_id" {}

# TODO: how to expose the app ID URI? azapi? (did via portal)
# TODO: how to add app ID URI to identifier URIs? (did via cli)
# az ad app update --id 074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f --identifier-uris api://074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f
resource "azuread_application" "dlqt_api" {
  display_name     = "dlqt-api"
  description      = "DLQT API"
  sign_in_audience = "AzureADMyOrg"
  owners           = ["f7ce87e4-54db-4a15-ae0e-e3fe0eef8eaa"] # me

  api {
    mapped_claims_enabled          = true
    requested_access_token_version = 2

    oauth2_permission_scope {
      value   = "dlq.read"
      type    = "User"
      id      = random_uuid.dlqt_api_scope_read_id.result
      enabled = true

      admin_consent_description  = "Read DLQ Messages"
      admin_consent_display_name = "Read DLQ Messages"
      user_consent_description   = "Read DLQ Messages"
      user_consent_display_name  = "Read DLQ Messages"
    }

    oauth2_permission_scope {
      value   = "dlq.retrigger"
      type    = "User"
      id      = random_uuid.dlqt_api_scope_retrigger_id.result
      enabled = true

      admin_consent_description  = "Retrigger DLQ Messages"
      admin_consent_display_name = "Retrigger DLQ Messages"
      user_consent_description   = "Retrigger DLQ Messages"
      user_consent_display_name  = "Retrigger DLQ Messages"
    }
  }

  lifecycle {
    ignore_changes = [ identifier_uris ]
  }
}

resource "azuread_service_principal" "dlqt_api" {
  client_id                    = azuread_application.dlqt_api.client_id
  app_role_assignment_required = true
}

resource "azuread_application_identifier_uri" "dlqt_api" {
  application_id = azuread_application.dlqt_api.id
  identifier_uri = "api://${azuread_application.dlqt_api.client_id}"
}

resource "azuread_application_pre_authorized" "dlqt_api" {
  application_id       = azuread_application.dlqt_api.id
  authorized_client_id = azuread_application.dlqt_cmd.client_id

  permission_ids = [
    resource.random_uuid.dlqt_api_scope_read_id.result,
    resource.random_uuid.dlqt_api_scope_retrigger_id.result,
  ]
}

# ===== DLQT CMD app reg ======

resource "azuread_application" "dlqt_cmd" {
  display_name     = "dlqt-cmd"
  description      = "DLQT CMD"
  sign_in_audience = "AzureADMyOrg"
  owners           = [local.my_principal_id]

  public_client {
    redirect_uris = ["http://localhost"] # allow AZ CLI
  }

  required_resource_access {
    resource_app_id = azuread_application.dlqt_api.client_id

    resource_access {
      id   = random_uuid.dlqt_api_scope_read_id.result
      type = "Scope"
    }

    resource_access {
      id   = random_uuid.dlqt_api_scope_retrigger_id.result
      type = "Scope"
    }
  }
}

resource "azuread_service_principal" "dlqt_cmd" {
  client_id                    = azuread_application.dlqt_cmd.client_id
  app_role_assignment_required = true
}
