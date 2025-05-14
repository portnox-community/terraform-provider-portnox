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
