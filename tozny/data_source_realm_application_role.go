package tozny

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// dataSourceRealmApplicationRole returns the schema and methods for provisioning a Tozny Realm Application Role
func dataSourceRealmApplicationRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRealmApplicationRoleRead,
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
				Computed:    true,
				ForceNew:    true,
			},
			"composite": {
				Description: "Whether this role is made of a combination of other roles.",
				Type: schema.TypeBool,
				Computed: true,
				ForceNew: true,
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

func dataSourceRealmApplicationRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

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

	d.Set("description", applicationRole.Description)
	d.Set("composite", applicationRole.Composite)
	d.Set("application_role_id", applicationRole.ID)
	d.SetId(applicationRole.ID)

	return diags
}
