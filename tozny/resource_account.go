package tozny

import (
  "context"

  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceAccount returns the schema and methods for provisioning a Tozny account.
func resourceAccount() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceAccountCreate,
    ReadContext:   resourceAccountRead,
    DeleteContext: resourceAccountDelete,
    Schema: map[string]*schema.Schema{
      "profile": {
        Description: "The account creator's profile settings.",
        Type:        schema.TypeList,
        MaxItems:    1,
        Required:    true,
        ForceNew:    true,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "id": {
              Description: "The unique user identifier for the profile.",
              Type:        schema.TypeInt,
              Computed:    false,
              Required:    true,
            },
          },
        },
      },
      "account": {
        Description: "Account wide settings.",
        Type:        schema.TypeList,
        MaxItems:    1,
        Required:    true,
        ForceNew:    true,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "id": {
              Description: "The unique user identifier for the account.",
              Type:        schema.TypeInt,
              Computed:    false,
              Required:    true,
            },
          },
        },
      },
    },
  }
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  // Warning or errors can be collected in a slice type
  var diags diag.Diagnostics

  return diags
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  // Warning or errors can be collected in a slice type
  var diags diag.Diagnostics

  return diags
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  // Warning or errors can be collected in a slice type
  var diags diag.Diagnostics

  return diags
}
