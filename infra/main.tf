resource "azurerm_resource_group" "this" {
  name     = "rg-dlqt"
  location = "centralus"
}

resource "azurerm_servicebus_namespace" "this" {
  name                = "sb-dlqt"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "Basic"
}

resource "azurerm_servicebus_queue" "this" {
  name         = "queue1"
  namespace_id = azurerm_servicebus_namespace.this.id
}
