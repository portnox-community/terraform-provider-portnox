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

```hcl
terraform {
  required_providers {
    portnox = {
      source  = "portnox-community/portnox"
      version = "1.0.0"
    }
  }
}

provider "portnox" {
  api_key = "your_api_key"
  retries = 100
}
```

Then, run:

```bash
terraform init
```

## Usage

### Example: Managing a MAC-Based Account

```hcl
resource "portnox_mac_account" "example" {
  account_name = "Example Account"
}
```

### Example: Managing Multiple MAC Addresses

```hcl
resource "portnox_mac_account_addresses" "example" {
  account_name = "Example Account"
  dynamic "mac_addresses" {
    for_each = var.mac_list
    content {
      mac_address = mac_addresses.value.mac_address
      description = mac_addresses.value.description
    }
  }
}
```

### Example: Managing an Individual MAC Address

```hcl
resource "portnox_mac_account_address" "example" {
  account_name = "Example Account"
  description  = "Example Description"
  mac_address  = "00:00:00:00:00:01"
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
