package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// dataSourceRealmApplication returns the schema and methods for gathering data about an existing Tozny Realm Application
func dataSourceRealmApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRealmApplicationRead,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the provider to use when provisioning this application.",
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
				Description: "The name of the Realm to provision the Application for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"client_id": {
				Description: "The external id for clients to reference when communicating with this application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_id": {
				Description: "Server defined unique identifier for the Application.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human readable/reference-able name for the application.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"protocol": {
				Description: "What protocol (e.g. OpenIDConnect or SAML) is used to authenticate with the application.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"active": {
				Description: "Whether this consumer is allowed to authenticate and authorize identities.",
				Type:        schema.TypeBool,
				Computed:    true,
				ForceNew:    true,
			},
			"oidc_settings": {
				Description: "Settings for an OIDC protocol based application.",
				Type:        schema.TypeList,
				Computed:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_origins": {
							Description: "The list of network locations that are allowed to be used by clients when accessing this application.",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
							ForceNew: true,
						},
						"access_type": {
							Description: "The OIDC access type.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
						"root_url": {
							Description: "The URL to append to any relative URLs.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
						"standard_flow_enabled": {
							Description: "Whether the OIDC standard flow is enabled",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"implicit_flow_enabled": {
							Description: "Whether the OIDC implicit flow is enabled",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"direct_access_grants_enabled": {
							Description: "Whether for OIDC flows direct access grants are enabled.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"base_url": {
							Description: "The OIDC base URL.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"saml_settings": {
				Description: "Settings for a SAML protocol based application.",
				Type:        schema.TypeList,
				Computed:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_origins": {
							Description: "The list of network locations that are allowed to be used by clients when accessing this application.",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
							ForceNew: true,
						},
						"default_endpoint": {
							Description: "URL used for every binding to both the SP's Assertion Consumer and Single Logout Services. This can be individually overridden for each binding and service.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
						"include_authn_statement": {
							Description: "Whether to include the Authn statement.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"include_one_time_use_condition": {
							Description: "Whether to include the one time use condition.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"sign_documents": {
							Description: "Whether to sign documents.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"sign_assertions": {
							Description: "Whether to sign assertions.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"client_signature_required": {
							Description: "Whether client signature is required.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"force_post_binding": {
							Description: "Whether to force POST binding.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"force_name_id_format": {
							Description: "Whether to force name ID format.",
							Type:        schema.TypeBool,
							Computed:    true,
							ForceNew:    true,
						},
						"name_id_format": {
							Description: "The name ID format",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
						"idp_initiated_sso_url_name": {
							Description: "The IDP initiated SSO URL name.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
						"assertion_consumer_service_post_binding_url": {
							Description: "The assertion consumer service post bind URL.",
							Type:        schema.TypeString,
							Computed:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRealmApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	realmName := d.Get("realm_name").(string)
	clientID := d.Get("client_id").(string)
	list, err := toznySDK.ListRealmApplications(ctx, realmName)
	if err != nil {
		return diag.FromErr(err)
	}

	var application *identityClient.Application
	for _, app := range list.Applications {
		if app.ClientID == clientID {
			application = &app
			break
		}
	}
	if application == nil {
		return diag.Errorf("Unable to find application with client id %q in realm %q", clientID, realmName)
	}

	d.Set("application_id", application.ID)
	d.Set("name", application.Name)
	d.Set("protocol", application.Protocol)
	d.Set("active", application.Active)

	// set SAML settings if this is a SAML application
	if application.Protocol == "openid-connect" {
		d.Set("oidc_settings", []interface{}{
			map[string]interface{}{
				"allowed_origins":              application.AllowedOrigins,
				"root_url":                     application.OIDCSettings.RootURL,
				"standard_flow_enabled":        application.OIDCSettings.StandardFlowEnabled,
				"implicit_flow_enabled":        application.OIDCSettings.ImplicitFlowEnabled,
				"direct_access_grants_enabled": application.OIDCSettings.DirectAccessGrantsEnabled,
				"base_url":                     application.OIDCSettings.BaseURL,
				"access_type":                  application.OIDCSettings.AccessType,
			},
		})
	} else if application.Protocol == "saml" {
		d.Set("saml_settings", []interface{}{
			map[string]interface{}{
				"allowed_origins":                             application.AllowedOrigins,
				"default_endpoint":                            application.SAMLSettings.DefaultEndpoint,
				"include_authn_statement":                     application.SAMLSettings.IncludeAuthnStatement,
				"include_one_time_use_condition":              application.SAMLSettings.IncludeOneTimeUseCondition,
				"sign_documents":                              application.SAMLSettings.SignDocuments,
				"sign_assertions":                             application.SAMLSettings.SignAssertions,
				"client_signature_required":                   application.SAMLSettings.ClientSignatureRequired,
				"force_post_binding":                          application.SAMLSettings.ForcePostBinding,
				"force_name_id_format":                        application.SAMLSettings.ForceNameIDFormat,
				"name_id_format":                              application.SAMLSettings.NameIDFormat,
				"idp_initiated_sso_url_name":                  application.SAMLSettings.IDPInitiatedSSOURLName,
				"assertion_consumer_service_post_binding_url": application.SAMLSettings.AssertionConsumerServicePOSTBindingURL,
			},
		})
	}

	// Associate created realm application  with Terraform state and signal success
	d.SetId(application.ID)

	return diags
}
