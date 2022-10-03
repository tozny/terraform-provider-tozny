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
			"alias": {
				Description: "User defined unique ID for the provider.",
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
						},
						"default_scope": {
							Description: "Default scope.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
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
	realmName := d.Get("realm_name").(string)
	config := d.Get("config").([]interface{})[0].(map[string]interface{})
	providerConfig := map[string]interface{}{
		"authorizationUrl": config["authorization_url"].(string),
		"tokenUrl":         config["token_url"].(string),
		"clientAuthMethod": config["client_auth_method"].(string),
		"clientId":         config["client_id"].(string),
		"clientSecret":     config["client_secret"].(string),
		"defaultScope":     config["default_scope"].(string),
	}
	createIdpRequest := identityClient.CreateIdentityProviderRequest{
		ProviderId:  "oidc",
		Alias:       d.Get("alias").(string),
		Config:      providerConfig,
		DisplayName: d.Get("display_name").(string),
		Enabled:     true,
	}
	err = toznySDK.CreateIdentityProvider(ctx, realmName, createIdpRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIdentityProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realmName := d.Get("realm_name").(string)
	alias := d.Get("alias").(string)
	err = toznySDK.DeleteIdentityProvider(ctx, realmName, alias)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
