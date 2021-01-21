package tozny

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmRole returns the schema and methods for provisioning a Tozny Realm Application Role
func resourceRealmRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmRoleCreate,
		ReadContext:   resourceRealmRoleRead,
		DeleteContext: resourceRealmRoleDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
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
				Description: "The name of the Realm to provision the realm Role for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human readable/reference-able name for the realm role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Human readable description for the realm role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_role_id": {
				Description: "Server defined unique identifier for the realm role.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"role_realm_id": {
				Description: "Server defined unique identifier for the realm associated with the role.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createRealmRoleParams := identityClient.CreateRealmRoleRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		Role: identityClient.Role{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
	}

	realmRole, err := toznySDK.CreateRealmRole(ctx, createRealmRoleParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("role_realm_id", realmRole.ContainerID)
	realmRoleID := realmRole.ID
	d.Set("realm_role_id", realmRoleID)

	d.SetId(realmRoleID)

	return diags
}

func resourceRealmRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	realmRole, err := toznySDK.DescribeRealmRole(ctx, identityClient.DescribeRealmRoleRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		RoleID:    d.Get("realm_role_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", realmRole.Name)
	d.Set("description", realmRole.Description)

	return diags
}

func resourceRealmRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmRole(ctx, identityClient.DeleteRealmRoleRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		RoleID:    d.Get("realm_role_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
