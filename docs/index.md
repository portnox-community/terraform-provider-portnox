# Documentation Index

Welcome to the documentation for the Terraform Provider for Portnox. Below is a list of available resources and data sources:

## Resources
- [MAC Account](resource_mac_account.md)
- [MAC Account Address](resource_mac_account_address.md)
- [MAC Account Addresses](resource_mac_account_addresses.md)

## Data Sources
- [MAC Account](datasource_mac_account.md)

## How to Use the Provider

To use the Portnox provider, include the following configuration in your Terraform script. Below is an example from `examples/main.tf`:

```hcl
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
  
  # Example of explicit declarations with validation
  mac_addresses {
      mac_address = "00:11:22:33:44:55"  # Must be in standard MAC format
      description = "networkprinter"      # Alphanumeric only, max 64 characters
  }
  
  # Example using dynamic blocks
  dynamic "mac_addresses" {
    for_each = var.mac_list
    content {
      mac_address = mac_addresses.value.mac_address  # Will be validated for proper MAC format
      description = mac_addresses.value.description  # Will be validated for alphanumeric, max 64 chars
      expiration  = mac_addresses.value.expiration
    }
  }
}

resource "portnox_mac_account_address" "example123" {
  account_name = "test"
  description  = "printer123"  # Alphanumeric only, max 64 characters
  mac_address  = "00:00:00:00:00:01"  # Must be in standard MAC format
  expiration   = "2025-12-31T23:59:59Z"
}
```

### Provider Specification

The `provider` block is used to configure the Portnox provider. Below is a breakdown of the key attributes:

- `api_key`: (Required) The API key used to authenticate with the Portnox API.
- `retries`: (Optional) The number of retry attempts for API requests. Default is `100`.

The `terraform` block specifies the required provider:

- `source`: The source of the provider, which is `portnox-community/portnox`.
- `version`: The version of the provider to use, e.g., `1.0.0`.

Example:

```hcl
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
```

Refer to the individual pages for detailed information on usage and examples.
