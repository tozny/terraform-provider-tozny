package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// dataSourceRealmRole returns the schema and methods for provisioning a Tozny Realm Application Role
func dataSourceRealmRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRealmRoleRead,
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
				Description: "Human readable description for the application role.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"composite": {
				Description: "Whether this role is made of a combination of other roles.",
				Type:        schema.TypeBool,
				Computed:    true,
				ForceNew:    true,
			},
			"realm_role_id": {
				Description: "Server defined unique identifier for the realm role.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func dataSourceRealmRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := d.Get("realm_name").(string)
	roleName := d.Get("name").(string)

	realmRoles, err := toznySDK.ListRealmRoles(ctx, realmName)
	var realmRole *identityClient.Role
	for _, role := range realmRoles.Roles {
		if role.Name == roleName {
			realmRole = &role
			break
		}
	}
	if realmRole == nil {
		return diag.Errorf("Unable to find realm role %q in realm %q", roleName, realmName)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", realmRole.Name)
	d.Set("description", realmRole.Description)
	d.Set("realm_role_id", realmRole.ID)
	d.SetId(realmRole.ID)

	return diags
}
