package tozny

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmApplicationRole returns the schema and methods for provisioning a Tozny Realm Application Role
func resourceRealmApplicationRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmApplicationRoleCreate,
		ReadContext:   resourceRealmApplicationRoleRead,
		DeleteContext: resourceRealmApplicationRoleDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the Realm to provision the Application Role for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_id": {
				Description: "Server defined unique identifier for the Application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human readable/reference-able name for the application role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Human readable description for the application role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_role_id": {
				Description: "Server defined unique identifier for the Application role.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmApplicationRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createApplicationRoleParams := identityClient.CreateRealmApplicationRoleRequest{
		RealmName:     strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID: d.Get("application_id").(string),
		ApplicationRole: identityClient.ApplicationRole{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
	}

	applicationRole, err := toznySDK.CreateRealmApplicationRole(ctx, createApplicationRoleParams)

	if err != nil {
		return diag.FromErr(err)
	}

	applicationRoleID := applicationRole.ID

	d.Set("application_role_id", applicationRoleID)
	d.SetId(applicationRoleID)

	return diags
}

func resourceRealmApplicationRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	applicationRole, err := toznySDK.DescribeRealmApplicationRole(ctx, identityClient.DescribeRealmApplicationRoleRequest{
		RealmName:           strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID:       d.Get("application_id").(string),
		ApplicationRoleName: d.Get("name").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", applicationRole.Name)
	d.Set("description", applicationRole.Description)

	return diags
}

func resourceRealmApplicationRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmApplicationRole(ctx, identityClient.DeleteRealmApplicationRoleRequest{
		RealmName:           strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID:       d.Get("application_id").(string),
		ApplicationRoleName: d.Get("name").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
