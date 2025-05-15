
## [1.0.4] - 2025-05-14
- Updated documentation for all resources to reflect validation rules for MAC addresses and description fields.
- Added more comprehensive examples in documentation to demonstrate proper usage with validations.

- Enhanced update logic to properly handle expiration changes for MAC addresses in the `portnox_mac_account_addresses` resource.

- Fixed issue where MAC addresses in `portnox_mac_account_addresses` resources were being reordered, causing unnecessary changes in Terraform plans.
- Modified the `Create`, `Read`, and `Update` lifecycle methods to preserve the original order of MAC addresses from the Terraform configuration.
- Added validation for MAC address format to ensure it follows the standard format (e.g., 00:00:00:00:00:00).
- Added validation for description field to limit it to 64 alphanumeric characters.

## [1.0.0] - 2025-05-14
- First stable release of the provider.
- Includes support for managing MAC-based accounts and their associated addresses.
- Added Terraform state management for MAC addresses.
