package tozny

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/tozny/e3db-clients-go/identityClient"
  "github.com/tozny/e3db-go/v2"
)

// resourceRealm returns the schema and methods for provisioning a Tozny Realm.
func resourceRealm() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceRealmCreate,
    ReadContext:   resourceRealmRead,
    DeleteContext: resourceRealmDelete,
    Schema: map[string]*schema.Schema{
      "realm_id": {
        Description: "Service defined unique identifier for the realm.",
        Type:        schema.TypeInt,
        Computed:    true,
        ForceNew:    true,
      },
      "domain": {
        Description: "Service defined & externally unique reference for the realm.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
      },
      "admin_url": {
        Description: "URL for realm administration console.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
      },
      "active": {
        Description: "Whether the realm is active for applications and identities to consume.",
        Type:        schema.TypeBool,
        Computed:    true,
        ForceNew:    true,
      },
      "broker_identity_tozny_id": {
        Description: "The Tozny Client ID associated with the Identity used to broker interactions between the realm and it's Identities. Will be empty if no realm broker Identity has been registered.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
      },
      "client_credentials_filepath": {
        Description: "The filepath to Tozny client credentials for the provider to use when provisioning this realm.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "realm_name": {
        Description: "User defined identifier for the realm.",
        Type:        schema.TypeString,
        Optional:    true,
        Computed:    true,
      },
      "sovereign_name": {
        Description: "User defined sovereign identifier.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "sovereign": {
        Description: "The admin identity for a realm.",
        Type:        schema.TypeList,
        Computed:    true,
        ForceNew:    true,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "id": {
              Description: "Service defined unique identifier for the sovereign.",
              Type:        schema.TypeInt,
              Computed:    true,
            },
            "name": {
              Description: "User defined sovereign identifier.",
              Type:        schema.TypeString,
              Optional:    true,
              Computed:    true,
            },
          },
        },
      },
    },
  }
}

func resourceRealmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznySDK := m.(*e3db.ToznySDKV3)

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)
  if toznyClientCredentialsFilePath != "" {
    toznySDK, err = e3db.GetSDKV3(toznyClientCredentialsFilePath)
    if err != nil {
      return diag.FromErr(err)
    }
  }

  realm, err := toznySDK.CreateRealm(ctx, identityClient.CreateRealmRequest{
    RealmName:     d.Get("realm_name").(string),
    SovereignName: d.Get("sovereign_name").(string),
  })

  if err != nil {
    return diag.FromErr(err)
  }

  d.Set("realm_id", realm.ID)
  d.Set("domain", realm.Domain)
  d.Set("admin_url", realm.AdminURL)
  d.Set("broker_identity_tozny_id", realm.BrokerIdentityToznyID)
  d.Set("active", realm.Active)
  d.Set("sovereign", []interface{}{
    map[string]interface{}{
      "id":   realm.Sovereign.ID,
      "name": realm.Sovereign.Name,
    },
  })

  // Associate created Realm with Terraform state and signal success
  d.SetId(fmt.Sprintf("%d", realm.ID))

  return diags
}

func resourceRealmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznySDK := m.(*e3db.ToznySDKV3)

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)
  if toznyClientCredentialsFilePath != "" {
    toznySDK, err = e3db.GetSDKV3(toznyClientCredentialsFilePath)
    if err != nil {
      return diag.FromErr(err)
    }
  }

  realm, err := toznySDK.DescribeRealm(ctx, d.Get("realm_name").(string))
  if err != nil {
    return diag.FromErr(err)
  }

  d.Set("realm_id", realm.ID)
  d.Set("domain", realm.Domain)
  d.Set("admin_url", realm.AdminURL)
  d.Set("broker_identity_tozny_id", realm.BrokerIdentityToznyID)
  d.Set("active", realm.Active)
  d.Set("sovereign", []interface{}{
    map[string]interface{}{
      "id":   realm.Sovereign.ID,
      "name": realm.Sovereign.Name,
    },
  })

  return diags
}

func resourceRealmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  var err error

  toznySDK := m.(*e3db.ToznySDKV3)

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)
  if toznyClientCredentialsFilePath != "" {
    toznySDK, err = e3db.GetSDKV3(toznyClientCredentialsFilePath)
    if err != nil {
      return diag.FromErr(err)
    }
  }

  err = toznySDK.DeleteRealm(ctx, d.Get("realm_name").(string))
  if err != nil {
    return diag.FromErr(err)
  }

  d.SetId("")

  return diags
}
