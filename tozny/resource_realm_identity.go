package tozny

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
	"github.com/tozny/e3db-go/v2"
)

// resourceRealmIdentity returns the schema and methods for configuring Tozny Identities
func resourceRealmIdentity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmIdentityCreate,
		ReadContext:   resourceRealmIdentityRead,
		DeleteContext: resourceRealmIdentityDelete,
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
				Description: "The name of the Realm to provision the identity for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"username": {
				Description: "The username for this identity",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"email": {
				Description: "The email address associated with this user.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"client_registration_token": {
				Description: "A registration token for the realm allowed to create identities",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"broker_target_url": {
				Description: "The base link for password resets",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Description: "The password for this identity. Ideally this comes from a secret store of some kind.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"first_name": {
				Description: "The first name associated with this identity",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"last_name": {
				Description: "The last name associated with this identity",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"recovery_email_ttl": {
				Description: "The length of time a recovery email is valid for",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmIdentityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := d.Get("realm_name").(string)
	registrationToken := d.Get("client_registration_token").(string)
	username := d.Get("username").(string)
	email := d.Get("email").(string)
	password := d.Get("password").(string)
	brokerTargetURL := d.Get("broker_target_url").(string)
	emailExpiryMinutes := d.Get("recovery_email_ttl").(int)
	realm := e3db.Realm{
		Name:               realmName,
		App:                "account",
		APIEndpoint:        toznySDK.APIEndpoint,
		BrokerTargetURL:    brokerTargetURL,
		EmailExpiryMinutes: emailExpiryMinutes,
	}

	identity, err := realm.Register(username, password, registrationToken, email, "", "")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(identity.ClientID)

	return diags
}

func resourceRealmIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// There is nothing to really update for an identity at this time, so implement as a noop.
	return diags
}

func resourceRealmIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := strings.ToLower(d.Get("realm_name").(string))

	err = toznySDK.DeleteIdentity(context.Background(), identityClient.RealmIdentityRequest{
		RealmName:  realmName,
		IdentityID: d.Id(),
	})
	if err != nil {
		diag.FromErr(fmt.Errorf("unable to delete identity: %+v", err))
	}

	d.SetId("")

	return diags
}
