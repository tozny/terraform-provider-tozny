package tozny

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmGroupRoleMappings returns the schema and methods for provisioning the role mappings for a Tozny Realm Group
func resourceRealmGroupRoleMappings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmGroupRoleMappingsCreate,
		ReadContext:   resourceRealmGroupRoleMappingsRead,
		DeleteContext: resourceRealmGroupRoleMappingsDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when provisioning role mappings for this realm group.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the Realm associated with the group to provision role mappings for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_id": {
				Description: "Server defined unique identifier for the group to provision role mappings for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_role": {
				Description: "Configuration for mapping an application role to members of a group.",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_id": {
							Description: "The application ID associated with the application role.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"role_id": {
							Description: "Service defined unique identifier for the application role.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"role_name": {
							Description: "User defined unique identifier for the application scoped role.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceRealmGroupRoleMappingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}
	// Attempt to add any role mappings specified by this resource
	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	if len(maybeTerraformApplicationRoleMappings) > 0 {
		roleMappings := groupRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)

		err = toznySDK.AddGroupRoleMappings(ctx, identityClient.RemoveGroupRoleMappingsRequest{
			RealmName:   strings.ToLower(d.Get("realm_name").(string)),
			GroupID:     d.Get("group_id").(string),
			RoleMapping: roleMappings,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}
	// The group role mapping lives on even if no role mappings exist
	// but we still need unique ID for Terraform's idempotent satisfaction
	d.SetId(uuid.New().String())

	return diags
}

func resourceRealmGroupRoleMappingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}

	groupRoleMappings, err := toznySDK.ListGroupRoleMappings(ctx, identityClient.ListGroupRoleMappingsRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		GroupID:   d.Get("group_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// Check whether we need to update state for this specific
	// group role mappings, which is only if one of the specified role mappings
	// don't exist on the server
	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	if len(maybeTerraformApplicationRoleMappings) > 0 {
		terraformGroupRoleMappings := groupRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)
		matchingGroupRoleMappings := identityClient.RoleMapping{
			ClientRoles: map[string][]identityClient.Role{},
		}
		var doUpdateState bool
		// Only check client application roles as that is currently
		// the only role type supported for provisioning with this resource
		// Group checking at a per application level so can skip any roles for applications this resource doesn't
		// map roles for
		for applicationID, terraformGroupApplicationRoleMappings := range terraformGroupRoleMappings.ClientRoles {
			applicationRoles, groupApplicationRoleMappingExists := groupRoleMappings.ClientRoles[applicationID]
			if !groupApplicationRoleMappingExists {
				doUpdateState = true
			}
			// Iterate over each specified client application role mapped for this group
			// and verify it exists on the server
			for _, terraformGroupApplicationRoleMapping := range terraformGroupApplicationRoleMappings {
				var matchingRoleFound bool
				for _, applicationRole := range applicationRoles {
					if applicationRole.ID == terraformGroupApplicationRoleMapping.ID {
						matchingRoleFound = true
						matchingGroupRoleMappings.ClientRoles[applicationRole.ContainerID] = append(matchingGroupRoleMappings.ClientRoles[applicationRole.ContainerID], applicationRole)
					}
				}
				// Role doesn't exist, need to update Terraform state
				if !matchingRoleFound {
					doUpdateState = true
				}
			}
		}
		if doUpdateState {
			var terraformRoleMappings []interface{}
			for applicationID, roleMappings := range matchingGroupRoleMappings.ClientRoles {
				for _, roleMapping := range roleMappings {
					terraformRoleMapping := map[string]interface{}{
						"application_id": applicationID,
						"role_id":        roleMapping.ID,
						"role_name":      roleMapping.Name,
					}
					terraformRoleMappings = append(terraformRoleMappings, terraformRoleMapping)
				}
			}
			d.Set("application_role", terraformRoleMappings)
		}
	}

	return diags
}

func resourceRealmGroupRoleMappingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

	if err != nil {
		return diag.FromErr(err)
	}
	// Attempt to delete any role mappings added by this resource
	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	if len(maybeTerraformApplicationRoleMappings) > 0 {
		roleMappings := groupRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)

		err = toznySDK.RemoveGroupRoleMappings(ctx, identityClient.RemoveGroupRoleMappingsRequest{
			RealmName:   strings.ToLower(d.Get("realm_name").(string)),
			GroupID:     d.Get("group_id").(string),
			RoleMapping: roleMappings,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}
	// The group role mapping lives on even if no role mappings exist
	// so we only need to cleanup the role mapping from terraform state
	d.SetId("")

	return diags
}

func groupRoleMappingsFromTerraform(data []interface{}) identityClient.RoleMapping {
	roleMappings := identityClient.RoleMapping{
		ClientRoles: map[string][]identityClient.Role{},
	}
	for _, terraformRoleMapping := range data {
		applicationRole := applicationRoleFromTerraform(terraformRoleMapping.(map[string]interface{}))
		roleMappings.ClientRoles[applicationRole.ContainerID] = append(roleMappings.ClientRoles[applicationRole.ContainerID], applicationRole)
	}
	return roleMappings
}

func applicationRoleFromTerraform(data map[string]interface{}) identityClient.Role {
	role := identityClient.Role{
		ID:          data["role_id"].(string),
		ContainerID: data["application_id"].(string),
		Name:        data["role_name"].(string),
		ClientRole:  true,
	}
	return role
}
