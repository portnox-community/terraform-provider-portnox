package portnox

import (
	"context"
	"log"
	"os"

	"github.com/portnox-community/terraform-provider-portnox/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns the schema.Provider for Portnox
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					value := os.Getenv("TF_VAR_PORTNOX_API_KEY")
					log.Printf("[DEBUG] Retrieved API Key: %s", value)
					return value, nil
				},
				Description: "The API key for accessing the Portnox API.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://clear.portnox.com:8081/CloudPortalBackEnd",
				Description: "The base URL for the Portnox API.",
			},
		},
		ResourcesMap:         map[string]*schema.Resource{},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	baseURL := d.Get("base_url").(string)

	// Allow apiKey to be populated by DefaultFunc if not explicitly set
	if apiKey == "" {
		return nil, diag.Errorf("API key must be provided either explicitly or via the PORTNOX_API_KEY environment variable")
	}

	logger := log.New(os.Stdout, "Portnox: ", log.LstdFlags)
	logger.Println("[DEBUG] Logger initialized and writing to stdout.")

	config := common.NewConfig(apiKey, baseURL, 3, 5, logger)

	return config, nil
}

type Config struct {
	APIKey  string
	BaseURL string
}
