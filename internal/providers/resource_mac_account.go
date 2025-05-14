package providers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMacAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacAccountCreate,
		ReadContext:   resourceMacAccountRead,
		DeleteContext: resourceMacAccountDelete,
		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the MAC-based account.",
				ForceNew:    true,
			},
			"block_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason the account is blocked.",
				ForceNew:    false,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation timestamp of the account.",
				ForceNew:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true, // Make this field read-only
				Description: "A description of the MAC-based account.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The group ID associated with the account.",
				ForceNew:    true, // Set ForceNew to true
			},
			"identity_type": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The identity type of the account.",
				ForceNew:    false,
			},
			"is_block_by_admin": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the account is blocked by an admin.",
				ForceNew:    false,
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization ID associated with the account.",
				ForceNew:    false,
			},
			"mac_whitelist": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The MAC address.",
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
				Description: "A list of MAC addresses in the whitelist with additional metadata.",
				Computed:    true, // Do not track changes to this field
			},
			"vendors_whitelist": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of vendor names in the whitelist.",
				ForceNew:    true, // Set ForceNew to true
			},
			"put_devices_into_voice_vlan": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether to put devices into the voice VLAN.",
				ForceNew:    true, // Set ForceNew to true
			},
			"identity_pre_shared_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The identity pre-shared key.",
				ForceNew:    true, // Set ForceNew to true
			},
		},
	}
}

func resourceMacAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	accountName := d.Get("account_name").(string)

	description := d.Get("description").(string)
	account := map[string]string{
		"AccountName": d.Get("account_name").(string),
	}
	if description != "" {
		account["Description"] = description
	}

	payload := map[string]interface{}{
		"MacBasedAccounts": []map[string]string{account},
	}

	// Process `mac_whitelist` blocks dynamically
	if v, ok := d.GetOk("mac_whitelist"); ok {
		macWhitelist := v.([]interface{})
		whitelistEntries := make([]map[string]interface{}, len(macWhitelist))
		for i, entry := range macWhitelist {
			entryMap := entry.(map[string]interface{})
			whitelistEntries[i] = map[string]interface{}{
				"Mac":         entryMap["mac"],
				"Description": entryMap["description"],
				"Expiration":  entryMap["expiration"],
			}
		}
		payload["MacWhiteList"] = whitelistEntries
	}

	// Ensure the POST request uses the base URL for the API endpoint
	endpoint := "/api/mac-based-accounts"

	if _, err := config.MakeRequestWithRetry("POST", endpoint, payload); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(accountName)

	return nil
}

func resourceMacAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)
	accountID := d.Id()

	responseBody, err := config.MakeRequestWithRetry("GET", "/api/mac-based-accounts/"+accountID, nil)
	if err != nil {
		// Attempt to parse the response body for specific error details
		var errorResponse struct {
			InternalErrorCode int    `json:"InternalErrorCode"`
			InternalError     string `json:"InternalError"`
		}
		if parseErr := json.Unmarshal(responseBody, &errorResponse); parseErr == nil {
			if errorResponse.InternalErrorCode == 5357 {
				log.Printf("[DEBUG] Account not found: %s", errorResponse.InternalError)
				log.Printf("[DEBUG] Clearing state for resource ID: %s", accountID)
				d.SetId("") // Clear the state to trigger recreation
				return diag.Diagnostics{
					diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "Resource not found",
						Detail:   "The resource is missing from the API and will be recreated on the next apply.",
					},
				} // Return a warning diagnostic to signal Terraform to recreate the resource
			}
		}

		// If parsing fails or the error is not specific, return the original error
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Account read response: %s", string(responseBody))

	// Parse JSON and populate Terraform state
	var account struct {
		AccountId        string `json:"AccountId"`
		AccountName      string `json:"AccountName"`
		AgentlessOptions struct {
			MacWhiteList []struct {
				Mac         string `json:"Mac"`
				Description string `json:"Description"`
				Expiration  string `json:"Expiration"`
			} `json:"MacWhiteList"`
		} `json:"AgentlessOptions"`
		// Add other fields as needed...
	}

	if err := json.Unmarshal(responseBody, &account); err != nil {
		return diag.FromErr(err)
	}

	d.Set("account_id", account.AccountId)
	d.Set("account_name", account.AccountName)
	// d.Set(...) for other fields

	// Ensure `mac_whitelist` is only set in the state if explicitly defined in the configuration
	if _, ok := d.GetOk("mac_whitelist"); ok {
		// Parse `mac_whitelist` blocks dynamically from the API response
		if len(account.AgentlessOptions.MacWhiteList) > 0 {
			whitelistEntries := make([]map[string]interface{}, len(account.AgentlessOptions.MacWhiteList))
			for i, entry := range account.AgentlessOptions.MacWhiteList {
				whitelistEntries[i] = map[string]interface{}{
					"mac":         entry.Mac,
					"description": entry.Description,
					"expiration":  entry.Expiration,
				}
			}
			d.Set("mac_whitelist", whitelistEntries)
		} else {
			d.Set("mac_whitelist", nil)
		}
	} else {
		// Clear `mac_whitelist` from the state if not explicitly defined
		d.Set("mac_whitelist", nil)
	}

	return nil
}

func resourceMacAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*common.Config)

	accountID := d.Id()

	if _, err := config.MakeRequestWithRetry("DELETE", "/api/mac-based-accounts/"+accountID, nil); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
