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
