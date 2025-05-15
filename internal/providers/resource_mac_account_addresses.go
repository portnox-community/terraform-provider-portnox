package providers

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"

	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceMacAccountAddresses() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacAccountAddressesCreate,
		ReadContext:   resourceMacAccountAddressesRead,
		UpdateContext: resourceMacAccountAddressesUpdate,
		DeleteContext: resourceMacAccountAddressesDelete,
		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the MAC-based account.",
				ForceNew:    true,
			},
			"mac_addresses": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of MAC addresses with descriptions.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"mac_address": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "The MAC address to be added to the whitelist.",
						ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`), "must be a valid MAC address format (e.g., 00:00:00:00:00:00)"),
					},
					"description": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "A description of the MAC address. Must be alphanumeric and maximum 64 characters.",
						ValidateFunc: validation.All(
							validation.StringLenBetween(0, 64),
							validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]*$`), "description must contain only alphanumeric characters and be up to 64 characters long"),
						),
					},
					"expiration": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The expiration date/time of the MAC address.",
					},
				},
				},
			},
		},
	}
}

// sortMacAddresses ensures consistent sorting of MAC addresses by mac_address first and then by description
// This function is used across Create, Read, and Update methods to maintain consistent ordering
func sortMacAddresses(macAddresses []interface{}) []interface{} {
	// Create a copy of the input slice to avoid modifying the original
	sortedMacAddresses := make([]map[string]interface{}, len(macAddresses))
	for i, mac := range macAddresses {
		if mac != nil {
			sortedMacAddresses[i] = mac.(map[string]interface{})
		}
	}

	// Sort MAC addresses by mac_address primarily and description secondarily
	sort.SliceStable(sortedMacAddresses, func(i, j int) bool {
		// Ensure both elements are valid
		if sortedMacAddresses[i] == nil || sortedMacAddresses[j] == nil {
			return false
		}

		// Get mac_address values safely
		macI, okI := sortedMacAddresses[i]["mac_address"].(string)
		macJ, okJ := sortedMacAddresses[j]["mac_address"].(string)

		// If either value is not a string or nil, use safe defaults
		if !okI || !okJ {
			return false
		}

		// Compare mac_addresses
		if macI == macJ {
			// If mac_addresses are equal, compare descriptions
			descI, okDescI := sortedMacAddresses[i]["description"].(string)
			descJ, okDescJ := sortedMacAddresses[j]["description"].(string)

			// If either description is not a string, use safe defaults
			if !okDescI || !okDescJ {
				return false
			}

			return descI < descJ
		}

		return macI < macJ
	})

	// Convert back to []interface{}
	sortedInterfaces := make([]interface{}, len(sortedMacAddresses))
	for i, mac := range sortedMacAddresses {
		sortedInterfaces[i] = mac
	}

	return sortedInterfaces
}

func resourceMacAccountAddressesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	payload := map[string]interface{}{
		"AccountName":  accountName,
		"MacWhiteList": []map[string]interface{}{},
	}

	// Store the original order of mac_addresses from the config
	originalMacOrder := make([]string, 0)

	if macAddresses, ok := d.GetOk("mac_addresses"); ok {
		// Preserve the original order from configuration
		for _, mac := range macAddresses.([]interface{}) {
			macMap := mac.(map[string]interface{})
			originalMacOrder = append(originalMacOrder, macMap["mac_address"].(string))

			entry := map[string]interface{}{
				"Mac":         macMap["mac_address"].(string),
				"Description": macMap["description"].(string),
			}
			if expiration, ok := macMap["expiration"].(string); ok && expiration != "" {
				entry["Expiration"] = expiration
			}
			payload["MacWhiteList"] = append(payload["MacWhiteList"].([]map[string]interface{}), entry)
		}
	}
	endpoint := "/api/mac-based-accounts/mac-whitelist-add"
	if _, err := config.MakeRequestWithRetry("POST", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(accountName)

	// Keep the original order in the state - this is important to avoid unnecessary changes
	if macAddresses, ok := d.GetOk("mac_addresses"); ok {
		d.Set("mac_addresses", macAddresses)
	}

	return nil
}

func resourceMacAccountAddressesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	// Store the original order of mac_addresses from the config
	originalMacOrder := make([]string, 0)
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			originalMacOrder = append(originalMacOrder, macMap["mac_address"].(string))
		}
	}

	macAddresses := make([]map[string]interface{}, 0)
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			entry := map[string]interface{}{
				"Description": macMap["description"].(string),
				"Mac":         macMap["mac_address"].(string),
			}
			if expiration, exists := macMap["expiration"].(string); exists && expiration != "" {
				entry["Expiration"] = expiration
			} else {
				entry["expiration"] = nil // Ensure the attribute is unset if no valid value exists
			}
			macAddresses = append(macAddresses, entry)
		}
	}

	payload := map[string]interface{}{
		"MacWhiteList": macAddresses,
	}

	// Fetch the current state from the API
	endpoint := "/api/mac-based-accounts/search"

	responseBytes, err := config.MakeRequestWithRetry("POST", endpoint, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Unmarshal the response into a map
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.FromErr(err)
	}

	// Parse the response to extract MAC whitelist items
	accounts := response["Accounts"].([]interface{})
	if len(accounts) == 0 {
		return diag.Errorf("No account found with name %s", accountName)
	}

	agentlessOptions := accounts[0].(map[string]interface{})["AgentlessOptions"].(map[string]interface{})
	macWhiteList := agentlessOptions["MacWhiteList"].(map[string]interface{})["_items"].([]interface{})

	// Prepare the list of MAC addresses to update the Terraform state
	macAddresses = make([]map[string]interface{}, 0) // Use '=' to update the existing variable
	// Filter MAC addresses to include only those defined in the current state or declared in the resource
	stateMacs := make(map[string]bool)
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			stateMacs[macMap["mac_address"].(string)] = true
		}
	}

	filteredMacAddresses := make([]map[string]interface{}, 0)
	for _, mac := range macWhiteList {
		if mac == nil {
			continue
		}
		macMap := mac.(map[string]interface{})
		macAddress := macMap["Mac"].(string)
		if !stateMacs[macAddress] {
			continue
		}
		entry := map[string]interface{}{
			"description": macMap["Description"].(string),
			"mac_address": macAddress,
		}
		if expiration, exists := macMap["Expiration"].(string); exists && expiration != "" {
			entry["expiration"] = expiration
		} else {
			entry["expiration"] = nil // Ensure the attribute is unset if no valid value exists
		}
		filteredMacAddresses = append(filteredMacAddresses, entry)
	}

	// Sort the MAC addresses by their mac_address and description fields to ensure consistent ordering
	sort.SliceStable(filteredMacAddresses, func(i, j int) bool {
		if filteredMacAddresses[i]["mac_address"].(string) == filteredMacAddresses[j]["mac_address"].(string) {
			return filteredMacAddresses[i]["description"].(string) < filteredMacAddresses[j]["description"].(string)
		}
		return filteredMacAddresses[i]["mac_address"].(string) < filteredMacAddresses[j]["mac_address"].(string)
	})
	// Create a map of mac_address to its data for easy lookup
	macAddressMap := make(map[string]map[string]interface{})
	for _, mac := range filteredMacAddresses {
		macAddressMap[mac["mac_address"].(string)] = mac
	}

	// Preserve the original order from configuration
	orderedMacAddresses := make([]interface{}, 0)

	// Use the original order if available, otherwise use sorted order
	if len(originalMacOrder) > 0 {
		for _, macAddr := range originalMacOrder {
			if mac, exists := macAddressMap[macAddr]; exists {
				orderedMacAddresses = append(orderedMacAddresses, mac)
				delete(macAddressMap, macAddr)
			}
		}
	}

	// If there are any MAC addresses that weren't in the original order, append them
	if len(macAddressMap) > 0 {
		for _, mac := range macAddressMap {
			orderedMacAddresses = append(orderedMacAddresses, mac)
		}
	}

	// Update the Terraform state with ordered MAC addresses (matching the configuration order)
	d.Set("mac_addresses", orderedMacAddresses)
	d.Set("account_name", accountName)
	return nil
}

func resourceMacAccountAddressesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	// Store the original order of mac_addresses from the config for later use
	originalMacOrder := make([]string, 0)
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			originalMacOrder = append(originalMacOrder, macMap["mac_address"].(string))
		}
	}

	// Prepare the current and updated lists of MAC addresses
	currentMacs := make(map[string]map[string]interface{})
	if old, _ := d.GetChange("mac_addresses"); old != nil {
		for _, mac := range old.([]interface{}) {
			macMap := mac.(map[string]interface{})
			currentMacs[macMap["mac_address"].(string)] = macMap
		}
	}

	// Get the updated MAC addresses (preserving the order from the config)
	updatedMacs := make(map[string]map[string]interface{})
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			updatedMacs[macMap["mac_address"].(string)] = macMap
		}
	}

	// Identify MAC addresses to remove
	for mac := range currentMacs {
		if _, exists := updatedMacs[mac]; !exists {
			payload := map[string]interface{}{
				"AccountName": accountName,
				"MacWhiteList": []map[string]interface{}{
					{"Mac": mac},
				},
			}
			endpoint := "/api/mac-based-accounts/mac-whitelist-remove"
			if _, err := config.MakeRequestWithRetry("DELETE", endpoint, payload); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// Identify MAC addresses with updated descriptions
	for mac, currentMac := range currentMacs {
		if updatedMac, exists := updatedMacs[mac]; exists {
			if currentMac["description"] != updatedMac["description"] {
				payload := map[string]interface{}{
					"AccountName": accountName,
					"MacWhiteList": []map[string]interface{}{
						{
							"Mac":         mac,
							"Description": updatedMac["description"],
						},
					},
				}
				endpoint := "/api/mac-based-accounts/mac-whitelist-remove"
				if _, err := config.MakeRequestWithRetry("DELETE", endpoint, payload); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	// Identify MAC addresses with updated expirations
	for mac, currentMac := range currentMacs {
		if updatedMac, exists := updatedMacs[mac]; exists {
			currentExpiration, currentHasExpiration := currentMac["expiration"].(string)
			updatedExpiration, updatedHasExpiration := updatedMac["expiration"].(string)
			
			// Check if expiration has changed
			if (currentHasExpiration != updatedHasExpiration) || (currentHasExpiration && updatedHasExpiration && currentExpiration != updatedExpiration) {
				payload := map[string]interface{}{
					"AccountName": accountName,
					"MacWhiteList": []map[string]interface{}{
						{
							"Mac": mac,
						},
					},
				}
				
				// Add expiration only if it exists
				if updatedHasExpiration && updatedExpiration != "" {
					payload["MacWhiteList"].([]map[string]interface{})[0]["Expiration"] = updatedExpiration
				}
				
				endpoint := "/api/mac-based-accounts/mac-whitelist-remove"
				if _, err := config.MakeRequestWithRetry("DELETE", endpoint, payload); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	// Prepare the payload with the updated list of MAC addresses to add or update
	macAddresses := make([]map[string]interface{}, 0)
	for _, macMap := range updatedMacs {
		entry := map[string]interface{}{
			"Mac":         macMap["mac_address"].(string),
			"Description": macMap["description"].(string),
		}
		if expiration, exists := macMap["expiration"].(string); exists && expiration != "" {
			entry["Expiration"] = expiration
		}
		macAddresses = append(macAddresses, entry)
	}

	payload := map[string]interface{}{
		"AccountName":  accountName,
		"MacWhiteList": macAddresses,
	}
	endpoint := "/api/mac-based-accounts/mac-whitelist-add"
	if _, err := config.MakeRequestWithRetry("POST", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}

	// Create a map of mac_address to its data for easy lookup
	macAddressMap := make(map[string]map[string]interface{})
	if macs, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macs.([]interface{}) {
			macMap := mac.(map[string]interface{})
			macAddressMap[macMap["mac_address"].(string)] = macMap
		}
	}

	// Preserve the original order from configuration
	orderedMacAddresses := make([]interface{}, 0)

	// Use the original order from the beginning of the Update function
	for _, macAddr := range originalMacOrder {
		if mac, exists := macAddressMap[macAddr]; exists {
			orderedMacAddresses = append(orderedMacAddresses, mac)
		}
	}

	// Update the Terraform state preserving the configuration's order
	d.Set("mac_addresses", orderedMacAddresses)
	d.Set("account_name", accountName)
	return nil
}

func resourceMacAccountAddressesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	payload := map[string]interface{}{
		"AccountName":  accountName,
		"MacWhiteList": []map[string]interface{}{},
	}

	if macAddresses, ok := d.GetOk("mac_addresses"); ok {
		for _, mac := range macAddresses.([]interface{}) {
			macMap := mac.(map[string]interface{})
			entry := map[string]interface{}{
				"Mac": macMap["mac_address"].(string),
			}
			payload["MacWhiteList"] = append(payload["MacWhiteList"].([]map[string]interface{}), entry)
		}
	}

	endpoint := "/api/mac-based-accounts/mac-whitelist-remove"
	if _, err := config.MakeRequestWithRetry("DELETE", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
