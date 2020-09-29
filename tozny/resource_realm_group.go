package tozny

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmGroup returns the schema and methods for provisioning a Tozny Realm Group
func resourceRealmGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmGroupCreate,
		ReadContext:   resourceRealmGroupRead,
		DeleteContext: resourceRealmGroupDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the Realm to provision the group for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human readable/reference-able name for the group.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_id": {
				Description: "Server defined unique identifier for the group.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createGroupParams := identityClient.CreateRealmGroupRequest{
		RealmName: d.Get("realm_name").(string),
		Group: identityClient.Group{
			Name: d.Get("name").(string),
		},
	}

	group, err := toznySDK.CreateRealmGroup(ctx, createGroupParams)

	if err != nil {
		return diag.FromErr(err)
	}

	groupId := group.ID

	d.Set("group_id", groupId)
	d.SetId(groupId)

	return diags
}

func resourceRealmGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	group, err := toznySDK.DescribeRealmGroup(ctx, identityClient.DescribeRealmGroupRequest{
		RealmName: d.Get("realm_name").(string),
		GroupID:   d.Get("group_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", group.Name)

	return diags
}

func resourceRealmGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmGroup(ctx, identityClient.DeleteRealmGroupRequest{
		RealmName: d.Get("realm_name").(string),
		GroupID:   d.Get("group_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
