package tozny

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceShadowRealmFederation returns the schema and methods for provisioning a Tozny Realm Provider.
func resourceShadowRealmFederation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceShadowRealmFederationCreate,
		ReadContext:   resourceShadowRealmFederationRead,
		DeleteContext: resourceShadowRealmFederationDelete,
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
			"federation_source": {
				Description: "The federation source for the provider. Defaults to `tozid`.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "tozid",
			},
			"realm_name": {
				Description: "User defined identifier for the current realm.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"primary_realm_name": {
				Description: "User defined identifier for the primary realm. Defaults to value for realm_name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"api_credential": {
				Description: "Server defined API Credential given by Primary Realm Federation initiation",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"primary_realm_endpoint": {
				Description: "Endpoint the Shadow Realm will use for communication to the Primary Realm",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"active": {
				Description: "Whether the provider is active. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"sync": {
				Description: "Whether the provider is enabled for syncing identities. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"sync_frequency": {
				Description: "How often the identities are being synced from the Primary instance of the federation in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"connection_id": {
				Description: "Server defined Unique Identitfier for a connection given by Primary Realm Federation initiation",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceShadowRealmFederationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("primary_realm_name").(string) == "" {
		d.Set("primary_realm_name", d.Get("realm_name").(string))
	}

	// Convert String to UUID
	connectionID := d.Get("connection_id").(string)

	params := identityClient.ConnectFederationRequest{
		RealmName:            d.Get("realm_name").(string),
		FederationSource:     d.Get("federation_source").(string),
		PrimaryRealmName:     d.Get("primary_realm_name").(string),
		Active:               d.Get("active").(bool),
		Sync:                 d.Get("sync").(bool),
		SyncFrequency:        d.Get("sync_frequency").(int),
		APICredential:        d.Get("api_credential").(string),
		PrimaryRealmEndpoint: d.Get("primary_realm_endpoint").(string),
		ConnectionID:         uuid.MustParse(connectionID),
	}

	err = toznySDK.ConfigureFederationConnection(ctx, params)
	if err != nil {
		return diag.FromErr(err)
	}
	// Associate created Realm Provider with Terraform state and signal success
	d.SetId(d.Get("connection_id").(string))

	return diags
}

func resourceShadowRealmFederationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceShadowRealmFederationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Currently no Remove federation endpoint
	d.SetId("")
	return diags
}
