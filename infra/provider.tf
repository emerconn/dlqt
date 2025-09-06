terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.43"
    }
  }

  backend "azurerm" {
    resource_group_name  = "rg-tfstate-dlqt"
    storage_account_name = "sttfstatedlqt"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}

provider "azurerm" {
  features {}
}
