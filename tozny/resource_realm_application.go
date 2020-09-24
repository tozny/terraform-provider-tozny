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
      "oidc_root_url": {
        Description: "The URL to append to any relative URLs.",
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

  createApplicationParams := identityClient.CreateRealmApplicationRequest{
    RealmName: d.Get("realm_name").(string),
    Application: identityClient.Application{
      ClientID: d.Get("client_id").(string),
      Name:     d.Get("name").(string),
      Active:   d.Get("active").(bool),
      Protocol: strings.ToLower(d.Get("protocol").(string)),
      OIDCSettings: identityClient.ApplicationOIDCSettings{
        RootURL: d.Get("oidc_root_url").(string),
      },
      SAMLSettings: identityClient.ApplicationSAMLSettings{
        DefaultSAMLEndpoint: d.Get("saml_endpoint").(string),
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

  d.Set("active", application.Active)

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
