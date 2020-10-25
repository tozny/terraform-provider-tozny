package tozny

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmApplication returns the schema and methods for provisioning a Tozny Realm Application
func resourceRealmApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmApplicationCreate,
		ReadContext:   resourceRealmApplicationRead,
		DeleteContext: resourceRealmApplicationDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the provider to use when provisioning this application.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the Realm to provision the Application for.",
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
			"client_id": {
				Description: "The external id for clients to reference when communicating with this application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human readable/reference-able name for the application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"protocol": {
				Description: "What protocol (e.g. OpenIDConnect or SAML) is used to authenticate with the application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"active": {
				Description: "Whether this consumer is allowed to authenticate and authorize identities.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"oidc_settings": {
				Description:   "Settings for an OIDC protocol based application.",
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"saml_settings"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_origins": {
							Description: "The list of network locations that are allowed to be used by clients when accessing this application.",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							ForceNew: true,
						},
						"access_type": {
							Description: "The OIDC access type.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Default:     "confidential",
						},
						"root_url": {
							Description: "The URL to append to any relative URLs.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"standard_flow_enabled": {
							Description: "Whether the OIDC standard flow is enabled",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							ForceNew:    true,
						},
						"base_url": {
							Description: "The OIDC base URL.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"saml_settings": {
				Description:   "Settings for a SAML protocol based application.",
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"oidc_settings"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_origins": {
							Description: "The list of network locations that are allowed to be used by clients when accessing this application.",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							ForceNew: true,
						},
						"default_endpoint": {
							Description: "URL used for every binding to both the SP's Assertion Consumer and Single Logout Services. This can be individually overridden for each binding and service.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"include_authn_statement": {
							Description: "Whether to include the Authn statement.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"include_one_time_use_condition": {
							Description: "Whether to include the one time use condition.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"sign_documents": {
							Description: "Whether to sign documents.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"sign_assertions": {
							Description: "Whether to sign assertions.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"client_signature_required": {
							Description: "Whether client signature is required.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"force_post_binding": {
							Description: "Whether to force POST binding.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"force_name_id_format": {
							Description: "Whether to force name ID format.",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"name_id_format": {
							Description: "The name ID format",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"idp_initiated_sso_url_name": {
							Description: "The IDP initiated SSO URL name.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"assertion_consumer_service_post_binding_url": {
							Description: "The assertion consumer service post bind URL.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceRealmApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createApplicationParams := identityClient.CreateRealmApplicationRequest{
		RealmName: d.Get("realm_name").(string),
		Application: identityClient.Application{
			ClientID: d.Get("client_id").(string),
			Name:     d.Get("name").(string),
			Active:   d.Get("active").(bool),
			Protocol: strings.ToLower(d.Get("protocol").(string)),
		},
	}

	maybeTerraformOIDCSettings := d.Get("oidc_settings").([]interface{})

	if len(maybeTerraformOIDCSettings) > 0 {
		terraformOIDCSettings := maybeTerraformOIDCSettings[0].(map[string]interface{})

		var allowedOrigins []string
		for _, allowedOrigin := range terraformOIDCSettings["allowed_origins"].([]interface{}) {
			allowedOrigins = append(allowedOrigins, allowedOrigin.(string))
		}
		createApplicationParams.Application.AllowedOrigins = allowedOrigins

		createApplicationParams.Application.OIDCSettings = identityClient.ApplicationOIDCSettings{
			RootURL:             terraformOIDCSettings["root_url"].(string),
			StandardFlowEnabled: terraformOIDCSettings["standard_flow_enabled"].(bool),
			BaseURL:             terraformOIDCSettings["base_url"].(string),
			AccessType:          terraformOIDCSettings["access_type"].(string),
		}
	}

	maybeTerraformSAMLSettings := d.Get("saml_settings").([]interface{})

	if len(maybeTerraformSAMLSettings) > 0 {
		terraformSAMLSettings := maybeTerraformSAMLSettings[0].(map[string]interface{})

		var allowedOrigins []string
		for _, allowedOrigin := range terraformSAMLSettings["allowed_origins"].([]interface{}) {
			allowedOrigins = append(allowedOrigins, allowedOrigin.(string))
		}
		createApplicationParams.Application.AllowedOrigins = allowedOrigins

		createApplicationParams.Application.SAMLSettings = identityClient.ApplicationSAMLSettings{
			DefaultEndpoint:                        terraformSAMLSettings["default_endpoint"].(string),
			IncludeAuthnStatement:                  terraformSAMLSettings["include_authn_statement"].(bool),
			IncludeOneTimeUseCondition:             terraformSAMLSettings["include_one_time_use_condition"].(bool),
			SignDocuments:                          terraformSAMLSettings["sign_documents"].(bool),
			SignAssertions:                         terraformSAMLSettings["sign_assertions"].(bool),
			ClientSignatureRequired:                terraformSAMLSettings["client_signature_required"].(bool),
			ForcePostBinding:                       terraformSAMLSettings["force_post_binding"].(bool),
			ForceNameIDFormat:                      terraformSAMLSettings["force_name_id_format"].(bool),
			NameIDFormat:                           terraformSAMLSettings["name_id_format"].(string),
			IDPInitiatedSSOURLName:                 terraformSAMLSettings["idp_initiated_sso_url_name"].(string),
			AssertionConsumerServicePOSTBindingURL: terraformSAMLSettings["assertion_consumer_service_post_binding_url"].(string),
		}
	}

	application, err := toznySDK.CreateRealmApplication(ctx, createApplicationParams)

	if err != nil {
		return diag.FromErr(err)
	}

	applicationID := application.ID

	d.Set("application_id", applicationID)

	// Associate created realm application  with Terraform state and signal success
	d.SetId(applicationID)

	return diags
}

func resourceRealmApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	application, err := toznySDK.DescribeRealmApplication(ctx, identityClient.DeleteRealmApplicationRequest{
		RealmName:     d.Get("realm_name").(string),
		ApplicationID: d.Get("application_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("client_id", application.ClientID)
	d.Set("name", application.Name)
	d.Set("active", application.Active)
	d.Set("protocol", application.Protocol)

	maybeTerraformSAMLSettings := d.Get("saml_settings").([]interface{})
	// only set oidc settings if no oidc_settings

	if len(maybeTerraformSAMLSettings) == 0 {
		d.Set("oidc_settings", []interface{}{
			map[string]interface{}{
				"allowed_origins":       application.AllowedOrigins,
				"root_url":              application.OIDCSettings.RootURL,
				"standard_flow_enabled": application.OIDCSettings.StandardFlowEnabled,
				"base_url":              application.OIDCSettings.BaseURL,
				"access_type":           application.OIDCSettings.AccessType,
			},
		})
	}

	maybeTerraformOIDCSettings := d.Get("oidc_settings").([]interface{})
	// only set saml settings if no oidc_settings

	if len(maybeTerraformOIDCSettings) == 0 {
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

	return diags
}

func resourceRealmApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmApplication(ctx, identityClient.DeleteRealmApplicationRequest{
		RealmName:     d.Get("realm_name").(string),
		ApplicationID: d.Get("application_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// Delete from Terraform state and signal success
	d.SetId("")

	return diags
}
