package tozny

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmBrokerIdentity returns the schema and methods for provisioning a Tozny Client Registration Token.
func resourceRealmBrokerIdentity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmBrokerIdentityCreate,
		ReadContext:   resourceRealmBrokerIdentityRead,
		DeleteContext: resourceRealmBrokerIdentityDelete,
		Schema: map[string]*schema.Schema{
			"persist_credentials_to": {
				Description:  "Where to persist the generated broker identity credentials. Default: file",
				Type:         schema.TypeString,
				Default:      "file",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"file", "terraform"}, false),
			},
			"client_registration_token": {
				Description: "Token to use when registering the Identity's client.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the Realm to register the brokering Identity for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "User defined name for the brokering Identity.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"broker_identity_credentials_save_filepath": {
				Description: "The filepath to persist the provisioned Identities credentials to.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"credentials": {
				Description: "The client credentials as a JSON string. Only populated when persist_credentails_to is set to 'terraform'",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the provider to use when provisioning this brokering Identity.",
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
			"identity_client_id": {
				Description: "Server defined unique identifier for the brokering Identity's client.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmBrokerIdentityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	clientRegistrationToken, realmName := d.Get("client_registration_token").(string), d.Get("realm_name").(string)
	persistTo := d.Get("persist_credentials_to").(string)

	brokerIdentityConfig := ToznyBrokerIdentityConfig{
		ClientRegistrationToken: clientRegistrationToken,
		Name:                    d.Get("name").(string),
		RealmName:               realmName,
	}

	brokerIdentity, secretKeys, err := MakeToznyBrokerIdentity(brokerIdentityConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	registeredBrokerIdentity, err := toznySDK.RegisterRealmBrokerIdentity(ctx, identityClient.RegisterRealmBrokerIdentityRequest{
		RealmName:              realmName,
		RealmRegistrationToken: clientRegistrationToken,
		Identity:               brokerIdentity,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	realmBrokerIdentityID := registeredBrokerIdentity.Identity.ToznyID.String()

	d.Set("identity_id", realmBrokerIdentityID)

	registeredBrokerIdentity.Identity.PrivateEncryptionKeys = map[string]string{
		secretKeys.PrivateEncryptionKey.Type: secretKeys.PrivateEncryptionKey.Material,
	}
	registeredBrokerIdentity.Identity.PrivateSigningKeys = map[string]string{
		secretKeys.PrivateSigningKey.Type: secretKeys.PrivateSigningKey.Material,
	}

	if persistTo == "file" {
		err = SaveToznyBrokerIdentity(d.Get("broker_identity_credentials_save_filepath").(string), registeredBrokerIdentity.Identity)

		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("credentials", "")
	} else {
		clientCredentialsJSONBytes, err := json.Marshal(registeredBrokerIdentity.Identity)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("credentials", string(clientCredentialsJSONBytes))
	}

	// Associate created realm broker identity with Terraform state and signal success
	d.SetId(realmBrokerIdentityID)

	return diags
}

func resourceRealmBrokerIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Update or delete for identities not currently supported by the Tozny API so no-op here
	var diags diag.Diagnostics

	return diags
}

func resourceRealmBrokerIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Soft delete (from Terraform state only) as Identity deletion is not currently supported by the Tozny API
	d.SetId("")

	return diags
}
