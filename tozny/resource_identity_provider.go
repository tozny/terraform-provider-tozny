package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmProvider returns the schema and methods for provisioning a Tozny Realm Provider.
func resourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityProviderCreate,
		ReadContext:   resourceIdentityProviderRead,
		DeleteContext: resourceIdentityProviderDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				ForceNew:      true,
				ConflictsWith: []string{"client_credentials_config"},
			},
			"client_credentials_config": {
				Description:   "The Tozny account client configuration as a JSON string",
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"client_credentials_filepath"},
			},
			"provider_id": {
				Description: "Service defined unique identifier for the provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the realm to associate the provider with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"display_name": {
				Description: "User defined name for the provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: "Whether the provider is enabled for syncing identities. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"config": {
				Description: "Settings for the identity provider.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_url": {
							Description: "Auth URL from azure.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"token_url": {
							Description: "Token URL from azure.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"client_id": {
							Description: "Client ID from azure.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"client_secret": {
							Description: "Client Secret from azure.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"client_auth_method": {
							Description: "Client Auth method to send token from TozID.",
							Type:        schema.TypeString,
							Default:     "client_secret_post",
							Optional:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceIdentityProviderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var err error
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	providerConfig := map[string]interface{}{
		"authorizationUrl": "https://test-auth-url.com",
		"tokenUrl":         "https://test-auth-url.com",
		"clientAuthMethod": "client_secret_post",
		"clientId":         "adadadadadadadadadadadadad",
		"clientSecret":     "aDAZCSFCCXVsdsdsffsfsfsff",
	}

	createIdpRequest := identityClient.CreateIdentityProviderRequest{
		ProviderId:  "oidc",
		Alias:       "az-testing-terraform",
		Config:      providerConfig,
		DisplayName: "Azure IdP Testing terraform",
		Enabled:     true,
	}

	err = toznySDK.CreateIdentityProvider(ctx, "localtest", createIdpRequest)

	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceIdentityProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
