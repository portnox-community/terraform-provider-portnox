package providers

import (
	"context"
	"encoding/json"
	"fmt"

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
			}, "mac_whitelist": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MAC address in the whitelist.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the MAC address.",
						},
						"expiration": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration date/time of the MAC address.",
						},
					},
				},
				Description: "A list of MAC addresses in the whitelist with their descriptions and expiration dates.",
			},
			"vendor_whitelist": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vendor_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the vendor.",
						},
						"vendor_prefixes": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of MAC address prefixes associated with this vendor.",
						},
					},
				},
				Description: "A list of vendors with their associated MAC address prefixes.",
			},
			"last_updated_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who last updated the account.",
			},
			"secure_mab_options": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Secure MAB options for the account.",
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
	d.Set("org_id", accountData["OrgId"]) // Parse AgentlessOptions
	if agentlessOptions, ok := accountData["AgentlessOptions"].(map[string]interface{}); ok {
		// Parse MacWhiteList with full details
		if macWhiteList, ok := agentlessOptions["MacWhiteList"].([]interface{}); ok {
			macDetailsList := make([]map[string]interface{}, 0)

			// Process each MAC address entry
			for _, item := range macWhiteList {
				if item == nil {
					continue
				}

				if macEntry, ok := item.(map[string]interface{}); ok {
					// Skip entries without a MAC address
					macAddress, hasMac := macEntry["Mac"].(string)
					if !hasMac || macAddress == "" {
						continue
					}

					// Create a new entry with standardized field names
					newEntry := map[string]interface{}{
						"mac_address": macAddress,
					}

					// Handle description (may be null)
					if desc, ok := macEntry["Description"].(string); ok {
						newEntry["description"] = desc
					} else {
						newEntry["description"] = ""
					}

					// Handle expiration (may be null)
					if exp, ok := macEntry["Expiration"].(string); ok && exp != "" {
						newEntry["expiration"] = exp
					} else {
						newEntry["expiration"] = ""
					}

					macDetailsList = append(macDetailsList, newEntry)
				}
			}

			if err := d.Set("mac_whitelist", macDetailsList); err != nil {
				return diag.Errorf("error setting mac_whitelist: %s", err)
			}
		}
		// Parse SecureMabOptions
		if secureMabOptions, ok := agentlessOptions["SecureMabOptions"].(map[string]interface{}); ok {
			secureMabMap := make(map[string]interface{})

			// Convert numeric values to strings for TypeMap
			if action, ok := secureMabOptions["Action"].(float64); ok {
				secureMabMap["action"] = fmt.Sprintf("%d", int(action))
			}

			if enabled, ok := secureMabOptions["Enabled"].(bool); ok {
				if enabled {
					secureMabMap["enabled"] = "true"
				} else {
					secureMabMap["enabled"] = "false"
				}
			}

			d.Set("secure_mab_options", secureMabMap)
		}

		// Parse VendorsWhiteList
		if vendorsWhiteList, ok := agentlessOptions["VendorsWhiteList"].([]interface{}); ok {
			vendorsList := make([]map[string]interface{}, 0, len(vendorsWhiteList))

			for _, vendor := range vendorsWhiteList {
				if vendorMap, ok := vendor.(map[string]interface{}); ok {
					newVendor := map[string]interface{}{}

					// Get vendor name
					if vendorName, ok := vendorMap["VendorName"].(string); ok {
						newVendor["vendor_name"] = vendorName
					} else {
						newVendor["vendor_name"] = ""
					}

					// Get vendor prefixes
					prefixesList := []string{}
					if prefixes, ok := vendorMap["VendorPrefixes"].([]interface{}); ok {
						for _, prefix := range prefixes {
							if prefixStr, ok := prefix.(string); ok {
								prefixesList = append(prefixesList, prefixStr)
							}
						}
					}
					newVendor["vendor_prefixes"] = prefixesList

					vendorsList = append(vendorsList, newVendor)
				}
			}

			if err := d.Set("vendor_whitelist", vendorsList); err != nil {
				return diag.Errorf("error setting vendor_whitelist: %s", err)
			}
		}

		// Get LastUpdatedBy
		if lastUpdatedBy, ok := accountData["LastUpdatedBy"].(string); ok && lastUpdatedBy != "" {
			d.Set("last_updated_by", lastUpdatedBy)
		} else {
			d.Set("last_updated_by", "")
		}
	}

	return nil
}
