package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmProvider returns the schema and methods for provisioning a Tozny Realm Provider.
func resourceIdentityProviderMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityProviderMapperCreate,
		ReadContext:   resourceIdentityProviderMapperRead,
		DeleteContext: resourceIdentityProviderMapperDelete,
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
			"alias": {
				Description: "User defined unique ID for the provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the mapper",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"identity_provider_mapper": {
				Description: "Mapper Type used to a specific user session attribute",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"config": {
				Description: "Settings for the identity provider mapper.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sync_mode": {
							Description: "Determines how the user attributes are synced.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"claim": {
							Description: "Denotes the role attribute to be used from the user session.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"claim_value": {
							Description: "Value configured for the role in External IdP.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"role": {
							Description: "Name of realm role to be mapped.",
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

func resourceIdentityProviderMapperCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	var err error
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realmName := d.Get("realm_name").(string)
	alias := d.Get("alias").(string)
	config := d.Get("config").([]interface{})[0].(map[string]interface{})
	providerMapperConfig := map[string]interface{}{
		"sync_mode":   config["sync_mode"].(string),
		"claim":       config["claim"].(string),
		"claim.value": config["claim_value"].(string),
		"role":        config["role"].(string),
	}
	createIdpMapperRequest := identityClient.IdentityProviderMapperRequest{
		Config:                 providerMapperConfig,
		Name:                   d.Get("name").(string),
		IdentityProviderAlias:  alias,
		IdentityProviderMapper: d.Get("identity_provider_mapper").(string),
	}
	err = toznySDK.CreateIdentityProviderMapper(ctx, realmName, alias, createIdpMapperRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceIdentityProviderMapperRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// toznySDK, err := MakeToznySDK(d, m)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// realmName := d.Get("realm_name").(string)
	// alias := d.Get("alias").(string)
	// idpRepresentation, err := toznySDK.GetIdentityProvider(ctx, realmName, alias)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// d.Set("alias", idpRepresentation.Alias)
	return diags
}

func resourceIdentityProviderMapperDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	// toznySDK, err := MakeToznySDK(d, m)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// realmName := d.Get("realm_name").(string)
	// alias := d.Get("alias").(string)
	// err = toznySDK.DeleteIdentityProvider(ctx, realmName, alias)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	return diags
}
