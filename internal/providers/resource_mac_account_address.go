package providers

import (
	"context"
	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMacAccountAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacAccountAddressCreate,
		ReadContext:   resourceMacAccountAddressRead,
		DeleteContext: resourceMacAccountAddressDelete,
		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the MAC-based account.",
				ForceNew:    true, // Ensure changes trigger recreation
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the MAC address.",
				ForceNew:    true, // Ensure changes trigger recreation
			},
			"mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAC address to be added to the whitelist.",
				ForceNew:    true, // Ensure changes trigger recreation
			},
			"expiration": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The expiration date/time of the MAC address.",
				ForceNew:    true, // Ensure changes trigger recreation
			},
		},
	}
}

func resourceMacAccountAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	accountName := d.Get("account_name").(string)
	macAddress := d.Get("mac_address").(string)
	description := d.Get("description").(string)
	expiration := d.Get("expiration").(string)

	payload := map[string]interface{}{
		"AccountName": accountName,
		"MacWhiteList": []map[string]interface{}{
			{
				"Description": description,
				"Mac":         macAddress,
			},
		},
	}

	// Add expiration to the payload only if it is specified
	if expiration != "" {
		payload["MacWhiteList"].([]map[string]interface{})[0]["Expiration"] = expiration
	}

	endpoint := "/api/mac-based-accounts/mac-whitelist-add"

	if _, err := config.MakeRequestWithRetry("POST", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(accountName + ":" + macAddress)

	return nil
}

func resourceMacAccountAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	macAddress := d.Get("mac_address").(string)
	description := d.Get("description").(string)
	expiration := d.Get("expiration").(string)

	payload := map[string]interface{}{
		"MacWhiteList": []map[string]interface{}{
			{
				"Description": description,
				"Mac":         macAddress,
				"Expiration":  expiration,
			},
		},
	}

	endpoint := "/api/mac-based-accounts/search"

	_, err := config.MakeRequestWithRetry("POST", endpoint, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Process the response and update the state
	d.Set("description", description)
	d.Set("mac_address", macAddress)
	d.Set("expiration", expiration)

	return nil
}

func resourceMacAccountAddressDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	accountName := d.Get("account_name").(string)
	macAddress := d.Get("mac_address").(string)
	description := d.Get("description").(string)
	expiration := d.Get("expiration").(string)

	payload := map[string]interface{}{
		"AccountName": accountName,
		"MacWhiteList": []map[string]interface{}{
			{
				"Description": description,
				"Mac":         macAddress,
			},
		},
	}

	// Add expiration to the payload only if it is specified
	if expiration != "" {
		payload["MacWhiteList"].([]map[string]interface{})[0]["Expiration"] = expiration
	}

	endpoint := "/api/mac-based-accounts/mac-whitelist-remove"

	if _, err := config.MakeRequestWithRetry("DELETE", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
