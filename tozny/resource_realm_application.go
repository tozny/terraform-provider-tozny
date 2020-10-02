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
      "allowed_origins": {
        Description: "The allowed origins for the application.",
        Type:        schema.TypeList,
        Elem: &schema.Schema{
          Type: schema.TypeString,
        },
        Optional: true,
        ForceNew: true,
      },
      "oidc_access_type": {
        Description: "The OIDC access type.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
        Default:     "confidential",
      },
      "oidc_root_url": {
        Description: "The URL to append to any relative URLs.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "oidc_standard_flow_enabled": {
        Description: "Whether the OIDC standard flow is enabled",
        Type:        schema.TypeBool,
        Optional:    true,
        Default:     true,
        ForceNew:    true,
      },
      "oidc_base_url": {
        Description: "The OIDC base URL.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_endpoint": {
        Description: "URL used for every binding to both the SP's Assertion Consumer and Single Logout Services. This can be individually overridden for each binding and service.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_include_authn_statement": {
        Description: "Whether to include the Authn statement.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_include_one_time_use_condition": {
        Description: "Whether to include the one time use condition.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_sign_documents": {
        Description: "Whether to sign documents.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_sign_assertions": {
        Description: "Whether to sign assertions.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_client_signature_required": {
        Description: "Whether client signature is required.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_force_post_binding": {
        Description: "Whether to force POST binding.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_force_name_id_format": {
        Description: "Whether to force name ID format.",
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_name_id_format": {
        Description: "The name ID format",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_idp_initiated_sso_url_name": {
        Description: "The IDP initiated SSO URL name.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "saml_assertion_consumer_service_post_binding_url": {
        Description: "The assertion consumer service post bind URL.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "application_id": {
        Description: "Server defined unique identifier for the Application.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
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

  var allowedOrigins []string

  for _, allowedOrigin := range d.Get("allowed_origins").([]interface{}) {
    allowedOrigins = append(allowedOrigins, allowedOrigin.(string))
  }

  createApplicationParams := identityClient.CreateRealmApplicationRequest{
    RealmName: d.Get("realm_name").(string),
    Application: identityClient.Application{
      ClientID: d.Get("client_id").(string),
      Name:     d.Get("name").(string),
      Active:   d.Get("active").(bool),
      Protocol: strings.ToLower(d.Get("protocol").(string)),
      AllowedOrigins: allowedOrigins,
      OIDCSettings: identityClient.ApplicationOIDCSettings{
        RootURL: d.Get("oidc_root_url").(string),
        StandardFlowEnabled: d.Get("oidc_standard_flow_enabled").(bool),
        BaseURL: d.Get("oidc_base_url").(string),
      },
      SAMLSettings: identityClient.ApplicationSAMLSettings{
        DefaultSAMLEndpoint: d.Get("saml_endpoint").(string),
        IncludeAuthnStatement: d.Get("saml_include_authn_statement").(bool),
        IncludeOneTimeUseCondition: d.Get("saml_include_one_time_use_condition").(bool),
        SignDocuments: d.Get("saml_sign_documents").(bool),
        SignAssertions: d.Get("saml_sign_assertions").(bool),
        ClientSignatureRequired: d.Get("saml_client_signature_required").(bool),
        ForcePostBinding: d.Get("saml_force_post_binding").(bool),
        ForceNameIDFormat: d.Get("saml_force_name_id_format").(bool),
        NameIDFormat: d.Get("saml_name_id_format").(string),
        IDPInitiatedSSOURLName: d.Get("saml_idp_initiated_sso_url_name").(string),
        AssertionConsumerServicePOSTBindingURL: d.Get("saml_assertion_consumer_service_post_binding_url").(string),
      },
    },
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
  d.Set("allowed_origins", application.AllowedOrigins)
  d.Set("protocol", application.Protocol)
  d.Set("oidc_root_url", application.OIDCSettings.RootURL)
  d.Set("oidc_standard_flow_enabled", application.OIDCSettings.StandardFlowEnabled)
  d.Set("oidc_base_url", application.OIDCSettings.BaseURL)
  d.Set("saml_endpoint", application.SAMLSettings.DefaultSAMLEndpoint)
  d.Set("saml_include_authn_statement", application.SAMLSettings.IncludeAuthnStatement)
  d.Set("saml_include_one_time_use_condition", application.SAMLSettings.IncludeOneTimeUseCondition)
  d.Set("saml_sign_documents", application.SAMLSettings.SignDocuments)
  d.Set("saml_sign_assertions", application.SAMLSettings.SignAssertions)
  d.Set("saml_client_signature_required", application.SAMLSettings.ClientSignatureRequired)
  d.Set("saml_force_post_binding", application.SAMLSettings.ForcePostBinding)
  d.Set("saml_force_name_id_format", application.SAMLSettings.ForceNameIDFormat)
  d.Set("saml_name_id_format", application.SAMLSettings.NameIDFormat)
  d.Set("saml_idp_initiated_sso_url_name", application.SAMLSettings.IDPInitiatedSSOURLName)
  d.Set("saml_assertion_consumer_service_post_binding_url", application.SAMLSettings.AssertionConsumerServicePOSTBindingURL)

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
