package providers

import (
	"context"

	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The MAC address to be added to the whitelist.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A description of the MAC address.",
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

func resourceMacAccountAddressesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	return nil
}

func resourceMacAccountAddressesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	// Prepare the payload with the list of MAC addresses
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
			}
			macAddresses = append(macAddresses, entry)
		}
	}

	payload := map[string]interface{}{
		"MacWhiteList": macAddresses,
	}

	endpoint := "/api/mac-based-accounts/search"

	_, err := config.MakeRequestWithRetry("POST", endpoint, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the Terraform state
	d.Set("mac_addresses", macAddresses)
	d.Set("account_name", accountName)
	return nil
}

func resourceMacAccountAddressesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountName := d.Get("account_name").(string)

	// Prepare the current and updated lists of MAC addresses
	currentMacs := make(map[string]map[string]interface{})
	if old, _ := d.GetChange("mac_addresses"); old != nil {
		for _, mac := range old.([]interface{}) {
			macMap := mac.(map[string]interface{})
			currentMacs[macMap["mac_address"].(string)] = macMap
		}
	}

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

	// Update the Terraform state
	d.Set("mac_addresses", macAddresses)
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
