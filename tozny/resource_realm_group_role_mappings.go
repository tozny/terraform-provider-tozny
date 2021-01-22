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
				Description:   "The filepath to Tozny client credentials for the Terraform provider to use when provisioning role mappings for this realm group.",
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
			"realm_role": {
				Description: "Configuration for mapping a realm role to members of a group.",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"realm_id": {
							Description: "The realm ID associated with the realm role.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"role_id": {
							Description: "Service defined unique identifier for the realm role.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"role_name": {
							Description: "User defined unique identifier for the realm scoped role.",
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

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}
	// Attempt to add any role mappings specified by this resource
	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	applicationRolesToMap := len(maybeTerraformApplicationRoleMappings) > 0
	var applicationRoleMappings map[string][]identityClient.Role
	if applicationRolesToMap {
		applicationRoleMappings = groupApplicationRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)
	}

	maybeTerraformRealmRoleMappings := d.Get("realm_role").([]interface{})
	realmRolesToMap := len(maybeTerraformRealmRoleMappings) > 0
	var realmRoleMappings []identityClient.Role
	if realmRolesToMap {
		realmRoleMappings = groupRealmRoleMappingsFromTerraform(maybeTerraformRealmRoleMappings)
	}

	if applicationRolesToMap || realmRolesToMap {
		addGroupRoleMappingsRequest := identityClient.AddGroupRoleMappingsRequest{
			RealmName:   strings.ToLower(d.Get("realm_name").(string)),
			GroupID:     d.Get("group_id").(string),
			RoleMapping: identityClient.RoleMapping{},
		}
		if applicationRolesToMap {
			addGroupRoleMappingsRequest.RoleMapping.ClientRoles = applicationRoleMappings
		}
		if realmRolesToMap {
			addGroupRoleMappingsRequest.RoleMapping.RealmRoles = realmRoleMappings
		}
		err = toznySDK.AddGroupRoleMappings(ctx, addGroupRoleMappingsRequest)
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

	toznySDK, err := MakeToznySDK(d, m)

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
	var doUpdateState bool
	matchingGroupRoleMappings := identityClient.RoleMapping{
		ClientRoles: map[string][]identityClient.Role{},
	}

	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	applicationRolesToMap := len(maybeTerraformApplicationRoleMappings) > 0
	if applicationRolesToMap {
		terraformApplicationGroupRoleMappings := groupApplicationRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)
		// Group checking at a per application level so can skip any roles for applications this resource doesn't
		// map roles for
		for applicationID, terraformGroupApplicationRoleMappings := range terraformApplicationGroupRoleMappings {
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
	}

	maybeTerraformRealmRoleMappings := d.Get("realm_role").([]interface{})
	realmRolesToMap := len(maybeTerraformRealmRoleMappings) > 0
	if realmRolesToMap {
		terraformRealmGroupRoleMappings := groupRealmRoleMappingsFromTerraform(maybeTerraformRealmRoleMappings)
		// Iterate over each specified realm role mapped for this group
		// and verify it exists on the server
		for _, terraformRealmGroupRoleMapping := range terraformRealmGroupRoleMappings {
			var matchingRoleFound bool
			for _, realmRole := range groupRoleMappings.RealmRoles {
				if realmRole.ID == terraformRealmGroupRoleMapping.ID {
					matchingRoleFound = true
					matchingGroupRoleMappings.RealmRoles = append(matchingGroupRoleMappings.RealmRoles, realmRole)
				}
			}
			// Role doesn't exist, need to update Terraform state
			if !matchingRoleFound {
				doUpdateState = true
			}
		}
	}

	if doUpdateState {
		var terraformApplicationRoleMappings []interface{}
		for applicationID, roleMappings := range matchingGroupRoleMappings.ClientRoles {
			for _, roleMapping := range roleMappings {
				terraformApplicationRoleMapping := map[string]interface{}{
					"application_id": applicationID,
					"role_id":        roleMapping.ID,
					"role_name":      roleMapping.Name,
				}
				terraformApplicationRoleMappings = append(terraformApplicationRoleMappings, terraformApplicationRoleMapping)
			}
		}
		d.Set("application_role", terraformApplicationRoleMappings)

		var terraformRealmRoleMappings []interface{}
		for _, roleMapping := range matchingGroupRoleMappings.RealmRoles {
			terraformRealmRoleMapping := map[string]interface{}{
				"realm_id":  roleMapping.ContainerID,
				"role_id":   roleMapping.ID,
				"role_name": roleMapping.Name,
			}
			terraformRealmRoleMappings = append(terraformRealmRoleMappings, terraformRealmRoleMapping)
		}
		d.Set("realm_role", terraformRealmRoleMappings)
	}
	return diags
}

func resourceRealmGroupRoleMappingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}
	// Attempt to delete any role mappings added by this resource
	maybeTerraformApplicationRoleMappings := d.Get("application_role").([]interface{})
	applicationRolesToMap := len(maybeTerraformApplicationRoleMappings) > 0
	var applicationRoleMappings map[string][]identityClient.Role
	if applicationRolesToMap {
		applicationRoleMappings = groupApplicationRoleMappingsFromTerraform(maybeTerraformApplicationRoleMappings)
	}

	maybeTerraformRealmRoleMappings := d.Get("realm_role").([]interface{})
	realmRolesToMap := len(maybeTerraformRealmRoleMappings) > 0
	var realmRoleMappings []identityClient.Role
	if realmRolesToMap {
		realmRoleMappings = groupRealmRoleMappingsFromTerraform(maybeTerraformRealmRoleMappings)
	}

	if applicationRolesToMap || realmRolesToMap {
		addGroupRoleMappingsRequest := identityClient.RemoveGroupRoleMappingsRequest{
			RealmName:   strings.ToLower(d.Get("realm_name").(string)),
			GroupID:     d.Get("group_id").(string),
			RoleMapping: identityClient.RoleMapping{},
		}
		if applicationRolesToMap {
			addGroupRoleMappingsRequest.RoleMapping.ClientRoles = applicationRoleMappings
		}
		if realmRolesToMap {
			addGroupRoleMappingsRequest.RoleMapping.RealmRoles = realmRoleMappings
		}
		err = toznySDK.RemoveGroupRoleMappings(ctx, addGroupRoleMappingsRequest)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// The group role mapping lives on even if no role mappings exist
	// so we only need to cleanup the role mapping from terraform state
	d.SetId("")

	return diags
}

func groupApplicationRoleMappingsFromTerraform(data []interface{}) map[string][]identityClient.Role {
	roleMappings := map[string][]identityClient.Role{}
	for _, terraformRoleMapping := range data {
		applicationRole := applicationRoleFromTerraform(terraformRoleMapping.(map[string]interface{}))
		roleMappings[applicationRole.ContainerID] = append(roleMappings[applicationRole.ContainerID], applicationRole)
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

func groupRealmRoleMappingsFromTerraform(data []interface{}) []identityClient.Role {
	var roleMappings []identityClient.Role
	for _, terraformRoleMapping := range data {
		realmRole := realmRoleFromTerraform(terraformRoleMapping.(map[string]interface{}))
		roleMappings = append(roleMappings, realmRole)
	}
	return roleMappings
}

func realmRoleFromTerraform(data map[string]interface{}) identityClient.Role {
	role := identityClient.Role{
		ID:          data["role_id"].(string),
		ContainerID: data["realm_id"].(string),
		Name:        data["role_name"].(string),
		ClientRole:  false,
	}
	return role
}
