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
				Description: "Filepath to Tozny client credentials in JSON format.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TOZNY_CLIENT_CREDENTIALS_FILEPATH", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tozny_account":                         resourceAccount(),
			"tozny_client_registration_token":       resourceClientRegistrationToken(),
			"tozny_realm_broker_identity":           resourceRealmBrokerIdentity(),
			"tozny_realm_broker_delegation":         resourceRealmBrokerDelegation(),
			"tozny_realm":                           resourceRealm(),
			"tozny_realm_role":                      resourceRealmRole(),
			"tozny_realm_application":               resourceRealmApplication(),
			"tozny_realm_application_mapper":        resourceRealmApplicationMapper(),
			"tozny_realm_application_client_secret": resourceRealmApplicationClientSecret(),
			"tozny_realm_application_role":          resourceRealmApplicationRole(),
			"tozny_realm_group":                     resourceRealmGroup(),
			"tozny_realm_group_role_mappings":       resourceRealmGroupRoleMappings(),
			"tozny_realm_provider":                  resourceRealmProvider(),
			"tozny_realm_provider_mapper":           resourceRealmProviderMapper(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tozny_realm_application_saml_description": dataSourceRealmApplicationSAMLDescription(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the Tozny provider for use in provisioning Tozny
// resources (accounts, clients, realms, identities, applications, groups, roles, etc...)
// initializing the ToznySDK with (in priority order) :
// 1.) Tozny client credentials from a user specified config file
// 2.) Account credentials (username & password) set on the provider that are used to derive key material
// for making account level requests and fetching the account queen client for other API service calls.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiEndpoint := d.Get("api_endpoint").(string)
	username := d.Get("account_username").(string)
	password := d.Get("account_password").(string)
	clientCredentialsFilepath := d.Get("tozny_credentials_json_filepath").(string)

	var sdkConfig e3db.ToznySDKConfig
	toznySDK := &e3db.ToznySDKV3{
		AccountUsername: username,
		AccountPassword: password,
		APIEndpoint:     apiEndpoint,
	}

	var terraformToznySDKResult TerraformToznySDKResult
	var err error
	// If specified parse and load client credentials from file
	if clientCredentialsFilepath != "" {
		sdkConfigFileJSON, err := e3db.LoadConfigFile(clientCredentialsFilepath)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		sdkConfig = e3db.ToznySDKConfig{
			ClientConfig: e3dbClients.ClientConfig{
				ClientID:  sdkConfigFileJSON.ClientID,
				APIKey:    sdkConfigFileJSON.APIKeyID,
				APISecret: sdkConfigFileJSON.APISecret,
				Host:      sdkConfigFileJSON.APIBaseURL,
				AuthNHost: sdkConfigFileJSON.APIBaseURL,
				SigningKeys: e3dbClients.SigningKeys{
					Public: e3dbClients.Key{
						Type:     e3dbClients.DefaultSigningKeyType,
						Material: sdkConfigFileJSON.PublicSigningKey,
					},
					Private: e3dbClients.Key{
						Type:     e3dbClients.DefaultSigningKeyType,
						Material: sdkConfigFileJSON.PrivateSigningKey,
					},
				},
				EncryptionKeys: e3dbClients.EncryptionKeys{
					Private: e3dbClients.Key{
						Material: sdkConfigFileJSON.PrivateKey,
						Type:     e3dbClients.DefaultEncryptionKeyType,
					},
					Public: e3dbClients.Key{
						Material: sdkConfigFileJSON.PublicKey,
						Type:     e3dbClients.DefaultEncryptionKeyType,
					},
				},
			},
			AccountUsername: sdkConfigFileJSON.AccountUsername,
			AccountPassword: sdkConfigFileJSON.AccountPassword,
			APIEndpoint:     sdkConfigFileJSON.APIBaseURL,
		}
	} else {
		// Otherwise attempt to derive credentials based off provider config
		// derive client credentials by logging in
		var accountConfig e3db.Account
		accountConfig, err = toznySDK.Login(ctx, username, password, "password", toznySDK.APIEndpoint)
		// Don't abort on error as valid provider config is optional or not desired for all resource use cases
		if err != nil {
			terraformToznySDKResult.SDK = toznySDK
			terraformToznySDKResult.Err = err
			return terraformToznySDKResult, diags
		}
		clientConfig := accountConfig.Config
		// seed sdk config with client credentials
		sdkConfig = e3db.ToznySDKConfig{
			ClientConfig: e3dbClients.ClientConfig{
				ClientID:  clientConfig.ClientID,
				APIKey:    clientConfig.APIKeyID,
				APISecret: clientConfig.APISecret,
				Host:      clientConfig.APIURL,
				AuthNHost: clientConfig.APIURL,
				SigningKeys: e3dbClients.SigningKeys{
					Public: e3dbClients.Key{
						Type:     e3dbClients.DefaultSigningKeyType,
						Material: clientConfig.PublicSigningKey,
					},
					Private: e3dbClients.Key{
						Type:     e3dbClients.DefaultSigningKeyType,
						Material: clientConfig.PrivateSigningKey,
					},
				},
				EncryptionKeys: e3dbClients.EncryptionKeys{
					Private: e3dbClients.Key{
						Material: clientConfig.PrivateKey,
						Type:     e3dbClients.DefaultEncryptionKeyType,
					},
					Public: e3dbClients.Key{
						Material: clientConfig.PublicKey,
						Type:     e3dbClients.DefaultEncryptionKeyType,
					},
				},
			},
			AccountUsername: username,
			AccountPassword: password,
			APIEndpoint:     clientConfig.APIURL,
		}

	}
	// Allow for overriding of any file based config via top level provider configuration
	if username != "" {
		sdkConfig.AccountUsername = username
	}
	if password != "" {
		sdkConfig.AccountPassword = password
	}
	if apiEndpoint != "" {
		sdkConfig.APIEndpoint = apiEndpoint
	}

	toznySDK, err = e3db.NewToznySDKV3(sdkConfig)

	if err != nil {
		terraformToznySDKResult.Err = err
		return terraformToznySDKResult, diag.FromErr(err)
	}
	terraformToznySDKResult.SDK = toznySDK
	terraformToznySDKResult.Err = nil
	return terraformToznySDKResult, diags
}

type TerraformToznySDKResult struct {
	SDK *e3db.ToznySDKV3
	Err error
}
