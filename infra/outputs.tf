output "dlqt_api_app_reg_id" {
  description = "app ID of the DLQT API app registration"
  value       = azuread_application.dlqt_api.client_id
}

output "dlqt_cmd_app_reg_id" {
  description = "app ID of the DLQT CMD app registration"
  value       = azuread_application.dlqt_cmd.client_id
}
