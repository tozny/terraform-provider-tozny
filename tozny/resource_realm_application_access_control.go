package tozny

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

func resourceRealmApplicationAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmApplicationAccessControlCreate,
		UpdateContext: resourceRealmApplicationAccessControlUpdate,
		DeleteContext: resourceRealmApplicationAccessControlDelete,
		ReadContext:   resourceRealmApplicationAccessControlRead,
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
				Description: "The name of the Realm associated with the application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_id": {
				Description: "Server defined unique identifier for the application.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: "Whether this application has managed access control.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"group": {
				Description: "Users within the selected groups can access this application.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Description: "Service defined unique identifier for the group.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"extend_to_children": {
							Description: "Wether SubGroups are able to access this application as well.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func resourceRealmApplicationAccessControlCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realmName := d.Get("realm_name").(string)
	applicationID := d.Get("application_id").(string)
	enabled := d.Get("enabled").(bool)

	// Set Access Control
	enableAccessControlParams := identityClient.AccessControlPolicyRequest{
		RealmName:     realmName,
		ApplicationID: applicationID,
		Enable:        enabled,
	}
	err = toznySDK.EnableAccessControlPolicy(ctx, enableAccessControlParams)
	if err != nil {
		return diag.FromErr(err)
	}
	// If Access Control was enabled we want to Add the Access Control Groups
	if enabled {
		maybeTerraformGroups := d.Get("group").([]interface{})
		accessControlGroupsExist := len(maybeTerraformGroups) > 0
		var accessControlGroups []identityClient.AccessControlPolicyGroup
		// Need to check if any groups were given
		if accessControlGroupsExist {
			accessControlGroups = accessControlGroupsFromTerraform(maybeTerraformGroups)
			addGroups := identityClient.AddAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        accessControlGroups,
			}
			err = toznySDK.AddAccessControlGroupsPolicy(ctx, addGroups)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// Associate created realm application  with Terraform state and signal success
	d.SetId(uuid.New().String())

	return diags
}

func resourceRealmApplicationAccessControlUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realmName := d.Get("realm_name").(string)
	applicationID := d.Get("application_id").(string)
	enabled := d.Get("enabled").(bool)
	// Check if Enabled has been updated
	if d.HasChange("enabled") {
		enableAccessControlParams := identityClient.AccessControlPolicyRequest{
			RealmName:     realmName,
			ApplicationID: applicationID,
			Enable:        enabled,
		}
		err = toznySDK.EnableAccessControlPolicy(ctx, enableAccessControlParams)
		if err != nil {
			return diag.FromErr(err)
		}
		// If it was changed to be enabled, we must add the groups
		if enabled {
			// If Access Control was enabled we want to Add the Access Control Groups
			maybeTerraformGroups := d.Get("group").([]interface{})
			accessControlGroupsExist := len(maybeTerraformGroups) > 0
			var accessControlGroups []identityClient.AccessControlPolicyGroup
			if accessControlGroupsExist {
				accessControlGroups = accessControlGroupsFromTerraform(maybeTerraformGroups)
			}

			addGroups := identityClient.AddAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        accessControlGroups,
			}
			err = toznySDK.AddAccessControlGroupsPolicy(ctx, addGroups)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// Check if the Groups have been updated
	if d.HasChange("group") {
		changedGroups, _ := d.GetChange("groups")
		previousGroups := changedGroups.([]interface{})
		newGroups := d.Get("group").([]interface{})
		lengthOfNewGroups := len(newGroups)
		lengthOfPreviousGroups := len(previousGroups)

		// We had no groups before, and now we have some
		if lengthOfPreviousGroups == 0 && lengthOfNewGroups != 0 {
			accessControlGroups := accessControlGroupsFromTerraform(newGroups)
			addGroups := identityClient.AddAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        accessControlGroups,
			}
			err = toznySDK.AddAccessControlGroupsPolicy(ctx, addGroups)
			if err != nil {
				return diag.FromErr(err)
			}
			return diags
		}
		// we had groups before and now we have none
		if lengthOfPreviousGroups != 0 && lengthOfNewGroups == 0 {
			accessControlGroups := accessControlGroupsFromTerraform(newGroups)
			removeGroups := identityClient.RemoveAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        accessControlGroups,
			}
			err = toznySDK.RemoveAccessControlGroupsPolicy(ctx, removeGroups)
			if err != nil {
				return diag.FromErr(err)
			}
			return diags
		}
		// We added/removed groups
		newAccessControlGroups := accessControlGroupsFromTerraform(newGroups)
		previousAccessControlGroups := accessControlGroupsFromTerraform(previousGroups)
		groupsToAdd, groupsToRemove := accessControlGroupsToChange(newAccessControlGroups, previousAccessControlGroups)
		// Verify we have groups to add
		if len(groupsToAdd) != 0 {
			addGroups := identityClient.AddAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        groupsToAdd,
			}
			err = toznySDK.AddAccessControlGroupsPolicy(ctx, addGroups)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		// Verify we have groups to remove
		if len(groupsToRemove) != 0 {
			removeGroups := identityClient.RemoveAccessControlPolicyGroupRequest{
				RealmName:     realmName,
				ApplicationID: applicationID,
				Groups:        groupsToRemove,
			}
			err = toznySDK.RemoveAccessControlGroupsPolicy(ctx, removeGroups)
			if err != nil {
				return diag.FromErr(err)
			}
		}

	}
	d.SetId(uuid.New().String())
	return diags
}

func resourceRealmApplicationAccessControlDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realmName := d.Get("realm_name").(string)
	applicationID := d.Get("application_id").(string)
	enabled := false
	// Disable Access Control Policy
	enableAccessControlParams := identityClient.AccessControlPolicyRequest{
		RealmName:     realmName,
		ApplicationID: applicationID,
		Enable:        enabled,
	}
	err = toznySDK.EnableAccessControlPolicy(ctx, enableAccessControlParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")

	return diags
}
func resourceRealmApplicationAccessControlRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	applicationID := d.Get("application_id").(string)
	realmName := d.Get("realm_name").(string)
	groups := d.Get("group").([]interface{})

	d.Set("realm_name", realmName)
	d.Set("application_id", applicationID)
	d.Set("group", groups)
	d.SetId(applicationID)
	return diags
}

func accessControlGroupsFromTerraform(data []interface{}) []identityClient.AccessControlPolicyGroup {
	var groups []identityClient.AccessControlPolicyGroup
	for _, terraformGroups := range data {
		group := groupFromTerraform(terraformGroups.(map[string]interface{}))
		groups = append(groups, group)
	}
	return groups
}

func groupFromTerraform(data map[string]interface{}) identityClient.AccessControlPolicyGroup {
	group := identityClient.AccessControlPolicyGroup{
		ID:               data["group_id"].(string),
		ExtendToChildren: data["extend_to_children"].(bool),
	}
	return group
}

func accessControlGroupsToChange(newGroupsList []identityClient.AccessControlPolicyGroup, oldGroupsList []identityClient.AccessControlPolicyGroup) ([]identityClient.AccessControlPolicyGroup, []identityClient.AccessControlPolicyGroup) {
	var groupsToRemove []identityClient.AccessControlPolicyGroup

	// Find Groups to Add/Remove
	for _, oldGroup := range oldGroupsList {
		found := false
		for _, newGroup := range newGroupsList {
			if oldGroup.ID == newGroup.ID {
				found = true
				continue
			}
		}
		if !found {
			groupsToRemove = append(groupsToRemove, oldGroup)
			continue
		}
	}
	return newGroupsList, groupsToRemove
}
