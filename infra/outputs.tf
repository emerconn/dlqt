output "app_reg_tenant_id" {
  description = "tenant ID of the Azure AD tenant"
  value       = data.azuread_client_config.current.tenant_id
}

output "app_reg_client_id_dlqt_api" {
  description = "app ID of the DLQT API app registration"
  value       = azuread_application.dlqt_api.client_id
}

output "app_reg_object_id_dlqt_api" {
  description = "object ID of the DLQT API app registration"
  value       = azuread_application.dlqt_api.object_id
}

output "app_reg_client_id_dlqt_cmd" {
  description = "app ID of the DLQT CMD app registration"
  value       = azuread_application.dlqt_cmd.client_id
}

output "app_reg_object_id_dlqt_cmd" {
  description = "object ID of the DLQT CMD app registration"
  value       = azuread_application.dlqt_cmd.object_id
}

output "container_app_log_stream_command" {
  description = "command to stream logs from the container app"
  value       = "az containerapp logs show -n ${azurerm_container_app.api.name} -g ${azurerm_resource_group.this.name} --follow"
}
