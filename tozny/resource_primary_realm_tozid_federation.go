package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourcePrimaryRealmFederation returns the schema and methods for provisioning a Tozny Federated Realm Provider.
func resourcePrimaryRealmFederation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrimaryRealmFederationCreate,
		ReadContext:   resourcePrimaryRealmFederationRead,
		DeleteContext: resourcePrimaryRealmFederationDelete,
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
				Optional:    true,
				Default:     "tozid",
				ForceNew:    true,
			},
			"realm_name": {
				Description: "Server defined Unique Identitfier for a connection given by Primary Realm Federation initiation",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"connection_id": {
				Description: "Server defined Unique Identitfier for a connection given by Primary Realm Federation initiation",
				Type:        schema.TypeString,
				ForceNew:    true,
				Computed:    true,
			},
			"api_credential": {
				Description: "Server defined API Credential given by Primary Realm Federation initiation",
				Type:        schema.TypeString,
				ForceNew:    true,
				Computed:    true,
			},
		},
	}
}

func resourcePrimaryRealmFederationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createProviderRequest := identityClient.InitializeFederationConnectionRequest{
		RealmName:        d.Get("realm_name").(string),
		FederationSource: d.Get("federation_source").(string),
	}

	initiateResponse, err := toznySDK.InitiateFederationConnection(ctx, createProviderRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	// Associate created Realm Provider with Terraform state and signal success
	d.SetId(initiateResponse.ConnectionID.String())

	// Set computed fields
	d.Set("connection_id", initiateResponse.ConnectionID.String())
	d.Set("api_credential", initiateResponse.APICredential)

	return diags
}

func resourcePrimaryRealmFederationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourcePrimaryRealmFederationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Currently no disable endpoint
	d.SetId("")
	return diags
}
