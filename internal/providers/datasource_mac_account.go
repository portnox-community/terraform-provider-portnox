package providers

import (
	"context"
	"encoding/json"

	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceMacAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMacAccountRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the MAC-based account.",
			},
			"account_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the MAC-based account.",
			},
			"block_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason the account is blocked.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation timestamp of the account.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of the MAC-based account.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The group ID associated with the account.",
			},
			"identity_type": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The identity type of the account.",
			},
			"is_block_by_admin": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the account is blocked by an admin.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization ID associated with the account.",
			},
			"mac_whitelist": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of MAC addresses in the whitelist.",
			},
		},
	}
}

func dataSourceMacAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	accountID := d.Get("account_id").(string)

	responseBody, err := config.MakeRequestWithRetry("GET", "/api/mac-based-accounts/"+accountID, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse the response and update the state
	var accountData map[string]interface{}
	// Replace json.NewDecoder with json.Unmarshal to handle []byte response
	if err := json.Unmarshal(responseBody, &accountData); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(accountID)
	d.Set("account_name", accountData["AccountName"])
	d.Set("block_reason", accountData["BlockReason"])
	d.Set("created_at", accountData["CreatedAt"])
	d.Set("description", accountData["Description"])
	d.Set("group_id", accountData["GroupId"])
	d.Set("identity_type", accountData["IdentityType"])
	d.Set("is_block_by_admin", accountData["IsBlockByAdmin"])
	d.Set("org_id", accountData["OrgId"])

	// Parse MacWhiteList
	if agentlessOptions, ok := accountData["AgentlessOptions"].(map[string]interface{}); ok {
		if macWhiteList, ok := agentlessOptions["MacWhiteList"].([]interface{}); ok {
			macs := make([]string, len(macWhiteList))
			for i, mac := range macWhiteList {
				if macEntry, ok := mac.(map[string]interface{}); ok {
					macs[i] = macEntry["Mac"].(string)
				}
			}
			d.Set("mac_whitelist", macs)
		}
	}

	return nil
}
