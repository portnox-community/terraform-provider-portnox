# Terraform Provider for Portnox

This repository contains the Terraform provider for managing resources in [Portnox](https://www.portnox.com/), a network access control solution. The provider allows users to manage MAC-based accounts, whitelist MAC addresses, and configure other Portnox resources programmatically using Terraform.

## Features

- **Resource Management**:
  - `portnox_mac_account`: Manage MAC-based accounts.
  - `portnox_mac_account_address`: Manage individual MAC addresses associated with accounts.
  - `portnox_mac_account_addresses`: Manage multiple MAC addresses in bulk.

- **Data Sources**:
  - `portnox_mac_account`: Retrieve information about existing MAC-based accounts.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.0.0 or later
- A valid Portnox API key

## Installation

To use this provider, add the following to your `main.tf` file:

```terraform
terraform {
  required_providers {
    portnox = {
      source  = "portnox-community/portnox"
      version = "1.0.0"
    }
  }
}

provider "portnox" {
  api_key = var.portnox_api_key
}
```

Then, run:

```bash
terraform init
```

## Usage

### Example: Managing a MAC-Based Account

```terraform
resource "portnox_mac_account" "example" {
  account_name = "Example Account"
  description  = "An example MAC-based account."
  group_id     = "12345"

  mac_whitelist = [
    {
      mac         = "00:11:22:33:44:55"
      description = "Example MAC"
      expiration  = "2025-12-31T23:59:59Z"
    }
  ]
}
```

### Example: Managing Multiple MAC Addresses

```terraform
resource "portnox_mac_account_addresses" "example" {
  account_name = "Example Account"

  mac_addresses = [
    {
      mac_address = "00:11:22:33:44:55"
      description = "Example MAC 1"
      expiration  = "2025-12-31T23:59:59Z"
    },
    {
      mac_address = "66:77:88:99:AA:BB"
      description = "Example MAC 2"
      expiration  = "2026-12-31T23:59:59Z"
    }
  ]
}
```

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) 1.20 or later
- [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk)

### Building the Provider

Run the following command to build the provider:

```bash
go build -o terraform-provider-portnox
```

### Testing

Run the following command to execute the tests:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

For support, please contact [Portnox Support](https://www.portnox.com/support/).
