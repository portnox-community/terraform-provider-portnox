provider "portnox" {
  api_key  = "your_api_key"
  retries  = 100
}

terraform {
  required_providers {
    portnox = {
      source = "portnox-community/portnox"
      version = "1.0.0"
    }
  }
}

resource "portnox_mac_account" "example123" {
  account_name = "test"
}

resource "portnox_mac_account_addresses" "example123" {
  account_name = "test"

  dynamic "mac_addresses" {
    for_each = var.mac_list
    content {
      mac_address = mac_addresses.value.mac_address
      description = mac_addresses.value.description
    }
  }
}

resource "portnox_mac_account_address" "example123" {
  account_name = "test"
  description  = "tes-"
  mac_address  = "00:00:00:00:00:01"
}
