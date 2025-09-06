output "dlqt_app_id" {
  description = "The Application ID of the DLQT app registration - use this as audience in auth middleware"
  value       = azuread_application.dlqt.client_id
}

output "dlqt_fetch_scope" {
  description = "The fetch scope to request tokens for - use this in CLI"
  value       = "${azuread_application.dlqt.client_id}/dlq.fetch"
}

output "dlqt_retrigger_scope" {
  description = "The retrigger scope to request tokens for - use this in CLI"
  value       = "${azuread_application.dlqt.client_id}/dlq.retrigger"
}

output "api_service_url" {
  description = "The URL of the deployed API service"
  value       = "https://${azurerm_container_app.api.latest_revision_fqdn}"
}
