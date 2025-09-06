output "dlqt_app_id" {
  description = "The Application ID of the DLQT app registration - use this as audience in auth middleware"
  value       = azuread_application.dlqt.client_id
}

output "dlqt_scope" {
  description = "The scope to request tokens for - use this in CLI"
  value       = "${azuread_application.dlqt.client_id}/.default"
}

output "auth_service_url" {
  description = "The URL of the deployed auth service"
  value       = "https://${azurerm_container_app.auth.latest_revision_fqdn}"
}
