package tozny

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealm returns the schema and methods for provisioning a Tozny Realm.
func resourceRealm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmCreate,
		ReadContext:   resourceRealmRead,
		DeleteContext: resourceRealmDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Description: "Service defined unique identifier for the realm.",
				Type:        schema.TypeInt,
				Computed:    true,
				ForceNew:    true,
			},
			"domain": {
				Description: "Service defined & externally unique reference for the realm.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"admin_url": {
				Description: "URL for realm administration console.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"active": {
				Description: "Whether the realm is active for applications and identities to consume.",
				Type:        schema.TypeBool,
				Computed:    true,
				ForceNew:    true,
			},
			"broker_identity_tozny_id": {
				Description: "The Tozny Client ID associated with the Identity used to broker interactions between the realm and it's Identities. Will be empty if no realm broker Identity has been registered.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the provider to use when provisioning this realm.",
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
				Description: "User defined identifier for the realm.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"default_registration_token": {
				Description: "The default registration token to use for registering new Identities with this Realm",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"sovereign_name": {
				Description: "User defined sovereign identifier.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"sovereign": {
				Description: "The admin identity for a realm.",
				Type:        schema.TypeList,
				Computed:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Service defined unique identifier for the sovereign.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"name": {
							Description: "User defined sovereign identifier.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"mpc_enabled": {
				Description: "Flag to enable MPC for the Realm.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"tozid_federation_enabled": {
				Description: "Flag to enable TozID Federation for the Realm.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	realm, err := toznySDK.CreateRealm(ctx, identityClient.CreateRealmRequest{
		RealmName:         d.Get("realm_name").(string),
		SovereignName:     d.Get("sovereign_name").(string),
		RegistrationToken: d.Get("default_registration_token").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	mpcEnabled := d.Get("mpc_enabled").(bool)
	tozidFederation := d.Get("tozid_federation_enabled").(bool)
	err = toznySDK.RealmSettingsUpdate(ctx, d.Get("realm_name").(string), identityClient.RealmSettingsUpdateRequest{
		MPCEnabled:             &mpcEnabled,
		TozIDFederationEnabled: &tozidFederation,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("realm_id", realm.ID)
	d.Set("domain", realm.Domain)
	d.Set("admin_url", realm.AdminURL)
	d.Set("broker_identity_tozny_id", realm.BrokerIdentityToznyID)
	d.Set("active", realm.Active)
	d.Set("sovereign", []interface{}{
		map[string]interface{}{
			"id":   realm.Sovereign.ID,
			"name": realm.Sovereign.Name,
		},
	})
	d.Set("mpc_enabled", d.Get("mpc_enabled"))
	d.Set("tozid_federation_enabled", d.Get("tozid_federation_enabled"))

	// Associate created Realm with Terraform state and signal success
	d.SetId(fmt.Sprintf("%d", realm.ID))

	return diags
}

func resourceRealmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	realm, err := toznySDK.DescribeRealm(ctx, d.Get("realm_name").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	privateRealmInfo, err := toznySDK.PrivateRealmInfo(ctx, d.Get("realm_name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("realm_id", realm.ID)
	d.Set("domain", realm.Domain)
	d.Set("admin_url", realm.AdminURL)
	d.Set("broker_identity_tozny_id", realm.BrokerIdentityToznyID)
	d.Set("active", realm.Active)
	d.Set("sovereign", []interface{}{
		map[string]interface{}{
			"id":   realm.Sovereign.ID,
			"name": realm.Sovereign.Name,
		},
	})
	d.Set("mpc_enabled", privateRealmInfo.MPCEnabled)
	d.Set("tozid_federation_enabled", privateRealmInfo.TozIDFederationEnabled)

	return diags
}

func resourceRealmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealm(ctx, d.Get("realm_name").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
