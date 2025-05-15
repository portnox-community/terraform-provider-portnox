
## [1.0.6] - 2025-05-14
- Added import functionality to `portnox_mac_account_addresses` resource to allow importing existing MAC accounts with their addresses.
- Enhanced error handling for imports to handle different API response formats.
- Updated import logic to handle both older and newer API response structures for the MAC whitelist.
- Added selective MAC address import feature that allows importing only specific MAC addresses by listing them in the import ID.
- Improved documentation to clarify that only MAC addresses explicitly declared in the resource configuration will be managed after import.

## [1.0.5] - 2025-05-14
- Fixed bug in `portnox_mac_account` data source where MAC whitelist and other fields were not being properly parsed from the API response.
- Improved error handling and type assertions in the data source to ensure robust parsing of API responses.
- Enhanced `portnox_mac_account` data source to include vendor MAC information with the addition of a `vendor_whitelist` field that exposes both vendor names and their associated MAC address prefixes.
- Enhanced `portnox_mac_account` data source to include detailed information for MAC addresses, including descriptions and expiration dates.
- Updated documentation for the data source to demonstrate accessing the detailed MAC address information.

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
