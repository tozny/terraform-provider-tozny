package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmProvider returns the schema and methods for provisioning a Tozny Realm Provider.
func resourceRealmProvider() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmProviderCreate,
		ReadContext:   resourceRealmProviderRead,
		DeleteContext: resourceRealmProviderDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"provider_id": {
				Description: "Service defined unique identifier for the provider.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the realm to associate the provider with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "User defined name for the provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"provider_type": {
				Description: "The type of provider. Valid values are `ldap`. Defaults to `ldap`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ldap",
				ForceNew:    true,
			},
			"active": {
				Description: "Whether the provider is enabled for syncing identities. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"import_identities": {
				Description: "If true, LDAP identities will be imported into the realm and synced via the configured sync policies. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"priority": {
				Description: "Priority for the provider when doing identity lookups for the realm. Lower numbers equal higher priority. Defaults to `0`.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				ForceNew:    true,
			},
			"connection_settings": {
				Description: "Settings for the realm to use when syncing identities from the provider.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "Type of the provider to connect to. Valid values are `ad` (Active Directory), `Red Hat Directory Server`, `Tivoli`, `Novell e Directory` or `other`.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"identity_name_attribute": {
							Description: "Name of LDAP attribute, which is mapped as the identity name. For many LDAP server vendors it can be 'uid'. For Active directory it can be 'sAMAccountName' or 'cn'. The attribute should be filled for all LDAP identity records you want to import from LDAP to the realm.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"rdn_attribute": {
							Description: "Name of LDAP attribute, which is used as RDN (top attribute) of typical user DN. Usually it's the same as Username LDAP attribute, however it's not required. For example for Active directory it's common to use 'cn' as RDN attribute when username attribute might be 'sAMAccountName'.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"uuid_attribute": {
							Description: "Name of LDAP attribute, which is used as unique object identifier (UUID) for objects in LDAP. For many LDAP server vendors it's 'entryUUID' however some are different. For example for Active directory it should be 'objectGUID'. If your LDAP server really doesn't support the notion of UUID, you can use any other attribute, which is supposed to be unique among LDAP users in tree. For example 'uid' or 'entryDN'.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"identity_object_classes": {
							Description: "All values of LDAP objectClass attribute for identities in LDAP. Newly created Realm identities will be written to LDAP with all those object classes and existing LDAP identity records are found just if they contain all those object classes.",
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"connection_url": {
							Description: "URL for connecting to provider.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"identity_dn": {
							Description: "Full DN of LDAP tree where your identities are. This DN is parent of LDAP identities. It could be for example 'ou=users,dc=example,dc=com' assuming that your typical identity will have DN like 'uid=john,ou=users,dc=example,dc=com'.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"authentication_type": {
							Description: "LDAP Authentication type. Valid values are 'none' (anonymous LDAP authentication) or 'simple' (Bind credential + Bind password authentication).",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"bind_dn": {
							Description: "DN of LDAP admin, which will be used by the Realm to access LDAP server.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"bind_credential": {
							Description: "Password of LDAP admin.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Sensitive:   true,
						},
						"search_scope": {
							Description: "For one level, we search for users just in DNs specified by Identity DNs. For subtree, we search in whole of their subtree. 1= `One Level` 2 = `Subtree`.",
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
						},
						"trust_store_spi_mode": {
							Description: "Specifies whether LDAP connection will use the truststore SPI with the truststore configured for the Realm. Valid values are `always`, `never`, or `ldapsOnly`.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"connection_pooling": {
							Description: "Specifies whether the realm use connection pooling for accessing LDAP server.",
							Type:        schema.TypeBool,
							Required:    true,
							ForceNew:    true,
						},
						"pagination": {
							Description: "Specifies whether the LDAP server to connect to supports pagination.",
							Type:        schema.TypeBool,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func fetchIdentityObjectClasses(terraformData *schema.ResourceData) []string {
	terraformConnectionSettings := terraformData.Get("connection_settings").([]interface{})[0].(map[string]interface{})

	terraformConnectionSettingsIdentityObjectClasses := terraformConnectionSettings["identity_object_classes"].([]interface{})

	connectionSettingsIdentityObjectClasses := []string{}

	for _, terraformConnectionSettingsIdentityObjectClasses := range terraformConnectionSettingsIdentityObjectClasses {
		connectionSettingsIdentityObjectClasses = append(connectionSettingsIdentityObjectClasses, terraformConnectionSettingsIdentityObjectClasses.(string))
	}

	return connectionSettingsIdentityObjectClasses
}

func resourceRealmProviderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	terraformConnectionSettings := d.Get("connection_settings").([]interface{})[0].(map[string]interface{})

	createProviderRequest := identityClient.CreateRealmProviderRequest{
		RealmName: d.Get("realm_name").(string),
		Provider: identityClient.Provider{
			Type:             d.Get("provider_type").(string),
			Name:             d.Get("name").(string),
			Active:           d.Get("active").(bool),
			ImportIdentities: d.Get("import_identities").(bool),
			Priority:         d.Get("priority").(int),
			ConnectionSettings: identityClient.ProviderConnectionSettings{
				Type:                  terraformConnectionSettings["type"].(string),
				IdentityNameAttribute: terraformConnectionSettings["identity_name_attribute"].(string),
				RDNAttribute:          terraformConnectionSettings["rdn_attribute"].(string),
				UUIDAttribute:         terraformConnectionSettings["uuid_attribute"].(string),
				IdentityObjectClasses: fetchIdentityObjectClasses(d),
				ConnectionURL:         terraformConnectionSettings["connection_url"].(string),
				IdentityDN:            terraformConnectionSettings["identity_dn"].(string),
				AuthenticationType:    terraformConnectionSettings["authentication_type"].(string),
				BindDN:                terraformConnectionSettings["bind_dn"].(string),
				BindCredential:        terraformConnectionSettings["bind_credential"].(string),
				SearchScope:           terraformConnectionSettings["search_scope"].(int),
				TrustStoreSPIMode:     terraformConnectionSettings["trust_store_spi_mode"].(string),
				ConnectionPooling:     terraformConnectionSettings["connection_pooling"].(bool),
				Pagination:            terraformConnectionSettings["pagination"].(bool),
			},
		},
	}

	provider, err := toznySDK.CreateRealmProvider(ctx, createProviderRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	providerID := provider.ID

	d.Set("provider_id", providerID)

	// Associate created Realm Provider with Terraform state and signal success
	d.SetId(providerID)

	return diags
}

func resourceRealmProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	describeProviderRequest := identityClient.DescribeRealmProviderRequest{
		RealmName:  d.Get("realm_name").(string),
		ProviderID: d.Get("provider_id").(string),
	}

	provider, err := toznySDK.DescribeRealmProvider(ctx, describeProviderRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	// to suppress spurious diffs only update the identity_object_classes if
	// there is an upstream change
	configuredIdentitityObjectClasses := fetchIdentityObjectClasses(d)

	actualIdentityObjectClasses := provider.ConnectionSettings.IdentityObjectClasses

	if len(configuredIdentitityObjectClasses) != len(actualIdentityObjectClasses) {
		configuredIdentitityObjectClasses = actualIdentityObjectClasses
	}

	for _, configuredIdentitityObjectClass := range configuredIdentitityObjectClasses {
		var match bool
		for _, actualIdentityObjectClass := range actualIdentityObjectClasses {
			if actualIdentityObjectClass == configuredIdentitityObjectClass {
				match = true
				break
			}
		}
		if !match {
			configuredIdentitityObjectClasses = actualIdentityObjectClasses
			break
		}
	}

	terraformConnectionSettings := d.Get("connection_settings").([]interface{})[0].(map[string]interface{})

	terraformConnectionSettingsBindCredential := terraformConnectionSettings["bind_credential"].(string)

	d.Set("provider_type", provider.Type)
	d.Set("name", provider.Name)
	d.Set("active", provider.Active)
	d.Set("import_identities", provider.ImportIdentities)
	d.Set("priority", provider.Priority)
	d.Set("connection_settings", []interface{}{
		map[string]interface{}{
			"type":                    provider.ConnectionSettings.Type,
			"identity_name_attribute": provider.ConnectionSettings.IdentityNameAttribute,
			"rdn_attribute":           provider.ConnectionSettings.RDNAttribute,
			"uuid_attribute":          provider.ConnectionSettings.UUIDAttribute,
			"identity_object_classes": configuredIdentitityObjectClasses,
			"connection_url":          provider.ConnectionSettings.ConnectionURL,
			"identity_dn":             provider.ConnectionSettings.IdentityDN,
			"authentication_type":     provider.ConnectionSettings.AuthenticationType,
			"bind_dn":                 provider.ConnectionSettings.BindDN,
			// API only returns ******* so even if the password change there is no way to know that
			"bind_credential":      terraformConnectionSettingsBindCredential,
			"search_scope":         provider.ConnectionSettings.SearchScope,
			"trust_store_spi_mode": provider.ConnectionSettings.TrustStoreSPIMode,
			"connection_pooling":   provider.ConnectionSettings.ConnectionPooling,
			"pagination":           provider.ConnectionSettings.Pagination,
		},
	})

	return diags
}

func resourceRealmProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	deleteRealmProviderParams := identityClient.DeleteRealmProviderRequest{
		RealmName:  d.Get("realm_name").(string),
		ProviderID: d.Get("provider_id").(string),
	}

	err = toznySDK.DeleteRealmProvider(ctx, deleteRealmProviderParams)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
