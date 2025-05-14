package provider

import (
	"context"
	"github.com/portnox-community/terraform-provider-portnox/common"
	providers "github.com/portnox-community/terraform-provider-portnox/internal/providers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns the schema.Provider for Portnox
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_VAR_PORTNOX_API_KEY", nil),
				Description: "The API key for accessing the Portnox API.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://clear.portnox.com:8081/CloudPortalBackEnd",
				Description: "The base URL for the Portnox API.",
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3, // Default number of retries
				Description: "The number of retries for API requests.",
			},
			"retry_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1, // Default retry interval in seconds
				Description: "The retry interval in seconds between retries.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"portnox_mac_account":           providers.ResourceMacAccount(),
			"portnox_mac_account_address":   providers.ResourceMacAccountAddress(),
			"portnox_mac_account_addresses": providers.ResourceMacAccountAddresses(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"portnox_mac_account": providers.DataSourceMacAccount(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			apiKey := d.Get("api_key").(string)
			baseURL := d.Get("base_url").(string)
			retries := d.Get("retries").(int)
			retryInterval := d.Get("retry_interval").(int)

			if apiKey == "" {
				return nil, diag.Errorf("API key must be provided")
			}

			return &common.Config{
				APIKey:        apiKey,
				BaseURL:       baseURL,
				Retries:       retries,
				RetryInterval: retryInterval,
			}, nil
		},
	}
}
