// Package tozny implements a Terraform Provider for automating provisioning of Tozny Services using Terraform.
package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go"
	"github.com/tozny/e3db-go/v2"
)

// Provider returns the Terraform schema for the Tozny provider for automating provisioning of Tozny Services using Terraform.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_endpoint": &schema.Schema{
				Description: "Network location for API management and provisioning of Tozny products & services.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://api.e3db.com",
			},
			"account_username": &schema.Schema{
				Description: "Tozny account username. Used to derive client credentials where appropriate.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TOZNY_ACCOUNT_USERNAME", nil),
			},
			"account_password": &schema.Schema{
				Description: "Tozny account password. Used to derive client credentials where appropriate.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TOZNY_ACCOUNT_PASSWORD", nil),
			},
			"tozny_credentials_json_filepath": &schema.Schema{
				Description: "Filepath to tozny client credentials in JSON format.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TOZNY_CLIENT_CREDENTIALS_FILEPATH", e3db.ProfileInterpolationConfigFilePath),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tozny_account": resourceAccount(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the Tozny provider for use in provisioning Tozny
// resources (accounts, clients, realms, identities, applications, groups, roles, etc...)
// initializing the ToznySDK with (in priority order) :
// 1.) Account credentials (username & password) set on the provider that are used to derive key material
// for making account level requests and fetching the account queen client for other API service calls.
// 2.) Tozny client credentials from a user specified config file
// 3.) Tozny client credentials from the standard Tozny config directory ~/.tozny
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Parse Tozny API endpoint and account and/or client credentials
	apiEndpoint := d.Get("api_endpoint").(string)
	username := d.Get("account_username").(string)
	password := d.Get("account_password").(string)
	// By default try to derive account (queen) client & credentials
	// for interacting with Tozny API services
	toznySDK, err := e3db.NewToznySDKV3(e3db.ToznySDKConfig{
		ClientConfig: e3dbClients.ClientConfig{
			Host:      apiEndpoint,
			AuthNHost: apiEndpoint,
		},
		AccountUsername: username,
		AccountPassword: password,
		APIEndpoint:     apiEndpoint,
	})
	if err != nil {
		return toznySDK, diag.FromErr(err)
	}
	return toznySDK, diags
}
