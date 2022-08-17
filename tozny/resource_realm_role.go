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
			"attribute": {
				Description: "Arbitrary string-string key value pairs.",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "The key to use for the attribute",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"values": {
							Description: "A list of string values",
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
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

	role := identityClient.Role{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	createRealmRoleParams := identityClient.CreateRealmRoleRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		Role:      role,
	}

	realmRole, err := toznySDK.CreateRealmRole(ctx, createRealmRoleParams)
	if err != nil {
		return diag.FromErr(err)
	}

	// Adding attributes (if specified) as part of the realm role creation
	realmRoleID := realmRole.ID
	if roleAttributes := attributesFromState(d); len(roleAttributes) != 0 {
		role.Attributes = roleAttributes
		updateRealmRoleParams := identityClient.UpdateRealmRoleRequest{
			RoleID:    realmRoleID,
			RealmName: strings.ToLower(d.Get("realm_name").(string)),
			Role:      role,
		}
		_, err = toznySDK.UpdateRealmRole(ctx, updateRealmRoleParams)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Set("role_realm_id", realmRole.ContainerID)
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

	attributes := attributesFromState(d)
	for key, value := range realmRole.Attributes {
		attributes[key] = value
	}
	d.Set("attribute", attributesToState(attributes))

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
