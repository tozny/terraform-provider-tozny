package tozny

import (
  "context"
  "fmt"
  "time"

  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/tozny/e3db-clients-go/accountClient"
)

// resourceClientRegistrationToken returns the schema and methods for provisioning a Tozny Client Registration Token.
func resourceClientRegistrationToken() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceClientRegistrationTokenCreate,
    ReadContext:   resourceClientRegistrationTokenRead,
    DeleteContext: resourceClientRegistrationTokenDelete,
    Schema: map[string]*schema.Schema{
      "name": {
        Description: "User defined identifier for the token.",
        Type:        schema.TypeString,
        Required:    true,
        ForceNew:    true,
      },
      "allowed_registration_client_types": {
        Description: "The client types that can be registered using the token. Valid types are `general`, `identity`, and `broker`",
        Type:        schema.TypeList,
        Required:    true,
        ForceNew:    true,
        MinItems:    1,
        Elem: &schema.Schema{
          Type:     schema.TypeString,
          ForceNew: true,
        },
      },
      "enabled": {
        Description: "Whether the clients can be registered using this token.",
        Type:        schema.TypeBool,
        Optional:    true,
        Default:     true,
        ForceNew:    true,
      },
      "one_time_use": {
        Description: "Whether the token is only valid for registering a single client.",
        Type:        schema.TypeBool,
        Optional:    true,
        Default:     false,
        ForceNew:    true,
      },
      "client_credentials_filepath": {
        Description: "The filepath to Tozny client credentials for the provider to use when provisioning this registration token.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "token": {
        Description: "Client registration token.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
        Sensitive:   true,
      },
    },
  }
}

func resourceClientRegistrationTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, account, err := MakeToznySession(ctx, toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  tokenName := d.Get("name").(string)
  tfAllowedTypes := d.Get("allowed_registration_client_types").([]interface{})
  allowedTypes := []string{}

  for _, tfAllowedType := range tfAllowedTypes {
    allowedTypes = append(allowedTypes, tfAllowedType.(string))
  }

  createTokenResponse, err := toznySDK.CreateRegistrationToken(ctx, accountClient.CreateRegistrationTokenRequest{
    AccountServiceToken: account.Token,
    Name:                tokenName,
    TokenPermissions: accountClient.TokenPermissions{
      Enabled:      d.Get("enabled").(bool),
      AllowedTypes: allowedTypes,
    },
  })

  if err != nil {
    return diag.FromErr(err)
  }

  clientRegistrationToken := createTokenResponse.Token

  d.Set("token", clientRegistrationToken)

  // Associate created token with Terraform state and signal success
  d.SetId(fmt.Sprintf("%d%v", time.Now().Unix(), tokenName))

  return diags
}

func resourceClientRegistrationTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, account, err := MakeToznySession(ctx, toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  listedRegistrationTokens, err := toznySDK.ListRegistrationTokens(ctx, account.Token)

  if err != nil {
    return diag.FromErr(err)
  }

  var listed bool

  for _, listedRegistrationToken := range *listedRegistrationTokens {
    if listedRegistrationToken.Name == d.Get("name").(string) {
      if listedRegistrationToken.Token == d.Get("token").(string) {
        listed = true

        d.Set("enabled", listedRegistrationToken.Permissions.Enabled)
        d.Set("allowed_registration_client_types", []interface{}{listedRegistrationToken.Permissions.AllowedTypes})

        break
      }
    }
  }

  if !listed {
    d.Set("token", "")
    d.Set("enabled", false)
  }

  return diags
}

func resourceClientRegistrationTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, account, err := MakeToznySession(ctx, toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  err = toznySDK.DeleteRegistrationToken(ctx, accountClient.DeleteRegistrationTokenRequest{
    Token:               d.Get("token").(string),
    AccountServiceToken: account.Token,
  })

  if err != nil {
    return diag.FromErr(err)
  }

  d.SetId("")

  return diags
}
