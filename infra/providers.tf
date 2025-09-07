terraform {
  required_version = ">= 1.0"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 3.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.43"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }

  backend "azurerm" {
    resource_group_name  = "rg-tfstate-dlqt"
    storage_account_name = "sttfstatedlqt"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}

provider "azuread" {
}

provider "azurerm" {
  features {}
  subscription_id = "506064be-2bb4-4ed5-b7c0-7cc6a6503dd1"
}
