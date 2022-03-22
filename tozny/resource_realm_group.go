package tozny

import (
	"context"
	"strings"

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
		UpdateContext: resourceRealmGroupUpdate,
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
			"group_id": {
				Description: "Server defined unique identifier for the group.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"access_policy": {
				Description: "The entities which control granting temporary access to groups through multi-party control.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_policy_id": {
							Description: "Server generated Id for the access policy",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"approval_role_ids": {
							Description: "The roles that can approve requests for groups with this access policy.",
							Type:        schema.TypeList,
							MinItems:    1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Required: true,
						},
						"required_approvals": {
							Description: "The number of approvals required for multi-party control of this group.",
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
						},
						"maximum_access_duration_seconds": {
							Description: "The maximum duration that access will last if approved, in seconds.",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"plugin_type": {
							Description: "The supported plugin type for the access policy (e.g. \"jira\").",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"plugin_id": {
							Description: "The ID of the plugin.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"plugin_mpc_flow_source": {
							Description: "The ID of the source that managed MPC (e.g. the ID of a Jira Board).",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// Converts the access policies resource from Terraform to a list of AccessPolicy
func accessPoliciesFromTerraform(data []interface{}) []identityClient.AccessPolicy {
	var policies []identityClient.AccessPolicy
	for _, terraformPolicy := range data {
		policy := accessPolicyFromTerraform(terraformPolicy.(map[string]interface{}))
		policies = append(policies, policy)
	}
	return policies
}

// Converts a single access policy argument from Terraform to the AccessPolicy API
func accessPolicyFromTerraform(data map[string]interface{}) identityClient.AccessPolicy {
	approvalRoles := approvalRolesFromTerraform(data["approval_role_ids"].([]interface{}))
	policy := identityClient.AccessPolicy{
		ID:                           int64(data["access_policy_id"].(int)),
		ApprovalRoles:                approvalRoles,
		RequiredApprovals:            data["required_approvals"].(int),
		MaximumAccessDurationSeconds: data["maximum_access_duration_seconds"].(int),
		PluginType:                   data["plugin_type"].(string),
		PluginID:                     data["plugin_id"].(string),
		PluginMPCFlowSource:          data["plugin_mpc_flow_source"].(string),
	}
	return policy
}

// Converts the approval roles argument from Terraform to a list of Roles
func approvalRolesFromTerraform(data []interface{}) []identityClient.Role {
	var roles []identityClient.Role
	for _, terraformApprovalRoleID := range data {
		role := identityClient.Role{
			ID: terraformApprovalRoleID.(string),
		}
		roles = append(roles, role)
	}
	return roles
}

// Converts a list of Role IDs to a list of Terraform approval role ids
func approvalRoleIdsToTerraform(roles []identityClient.Role) []interface{} {
	var ids []interface{}
	for _, role := range roles {
		ids = append(ids, role.ID)
	}
	return ids
}

// Iterates through the API's access policies and maps the values to a structure that Terraform can assign
// to the Group resource's access_policy argument
func flattenAccessPolicyItems(groupAccessPolicy identityClient.GroupAccessPolicies) []interface{} {
	// If no access policies, return empty interface
	if len(groupAccessPolicy.AccessPolicies) == 0 {
		return make([]interface{}, 0)
	}

	// Build each flattened policy and append to access policies
	accessPolicies := make([]interface{}, len(groupAccessPolicy.AccessPolicies), len(groupAccessPolicy.AccessPolicies))
	for i, policy := range groupAccessPolicy.AccessPolicies {
		accessPolicy := make(map[string]interface{})
		approvalRoleIds := approvalRoleIdsToTerraform(policy.ApprovalRoles)
		accessPolicy["access_policy_id"] = policy.ID
		accessPolicy["approval_role_ids"] = approvalRoleIds
		accessPolicy["required_approvals"] = policy.RequiredApprovals
		accessPolicy["maximum_access_duration_seconds"] = policy.MaximumAccessDurationSeconds
		accessPolicy["plugin_type"] = policy.PluginType
		accessPolicy["plugin_id"] = policy.PluginID
		accessPolicy["plugin_mpc_flow_source"] = policy.PluginMPCFlowSource

		accessPolicies[i] = accessPolicy
	}

	return accessPolicies
}

func resourceRealmGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createGroupParams := identityClient.CreateRealmGroupRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		Group: identityClient.Group{
			Name:       d.Get("name").(string),
			Attributes: attributesFromState(d),
		},
	}

	group, err := toznySDK.CreateRealmGroup(ctx, createGroupParams)

	if err != nil {
		return diag.FromErr(err)
	}

	groupID := group.ID

	d.Set("group_id", groupID)
	d.SetId(groupID)

	// Create and upsert access policy attached to group, if one exists
	resourceAccessPolicies := d.Get("access_policy").([]interface{})
	accessPoliciesExist := len(resourceAccessPolicies) > 0
	if accessPoliciesExist {
		accessPolicies := accessPoliciesFromTerraform(resourceAccessPolicies)

		groupAccessPolicies := identityClient.GroupAccessPolicies{
			GroupID:        d.Get("group_id").(string),
			AccessPolicies: accessPolicies,
		}
		accessPolicyParams := identityClient.UpsertAccessPolicyRequest{
			RealmName:           d.Get("realm_name").(string),
			GroupAccessPolicies: groupAccessPolicies,
		}

		// Need to set the computed ID of the Access Policy
		policies, err := toznySDK.UpsertAccessPolicies(ctx, accessPolicyParams)
		if err != nil {
			return diag.FromErr(err)
		}

		accessPoliciesForTerraform := flattenAccessPolicyItems(policies.GroupAccessPolicies) // Only one Group's access policies were requested
		if err := d.Set("access_policy", accessPoliciesForTerraform); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRealmGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	group, err := toznySDK.DescribeRealmGroup(ctx, identityClient.DescribeRealmGroupRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		GroupID:   d.Get("group_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}
	// Update attributes
	attributes := attributesFromState(d)
	for key, value := range group.Attributes {
		attributes[key] = value
	}
	d.Set("attribute", attributesToState(attributes))

	d.Set("name", group.Name)

	// Only reading access policies attached to this group
	groupIDs := []string{d.Get("group_id").(string)}
	listAccessPoliciesRequest := identityClient.ListAccessPoliciesRequest{
		RealmName: d.Get("realm_name").(string),
		GroupIDs:  groupIDs,
	}

	realmAccessPolicies, err := toznySDK.ListAccessPolicies(ctx, listAccessPoliciesRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	// Access policies for only a single group were requested (this resource), so flatten the only element in the response
	accessPolicies := flattenAccessPolicyItems(realmAccessPolicies.GroupAccessPolicies[0])
	if err := d.Set("access_policy", accessPolicies); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRealmGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	// Semantically, deleting an access policy means to set the list of access policies attached to the
	// group to an empty list. Thereby removing an access policies associated with the group.
	groupAccessPolicies := identityClient.GroupAccessPolicies{
		GroupID:        d.Get("group_id").(string),
		AccessPolicies: []identityClient.AccessPolicy{}, // An empty list of AccessPolicy
	}
	accessPolicyParams := identityClient.UpsertAccessPolicyRequest{
		RealmName:           d.Get("realm_name").(string),
		GroupAccessPolicies: groupAccessPolicies,
	}

	// Nothing to do in the case of a successful access policy upsert. It simply returns the group's access policies.
	_, err = toznySDK.UpsertAccessPolicies(ctx, accessPolicyParams)
	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmGroup(ctx, identityClient.DeleteRealmGroupRequest{
		RealmName: strings.ToLower(d.Get("realm_name").(string)),
		GroupID:   d.Get("group_id").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceRealmGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Group update only handles updates to access policies.
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if any relevant properties have changed
	if d.HasChanges("access_policy") {
		// Update the access policy. Currently we only update plugin information.
		resourceAccessPolicies := d.Get("access_policy").([]interface{})
		accessPolicies := accessPoliciesFromTerraform(resourceAccessPolicies)

		groupAccessPolicies := identityClient.GroupAccessPolicies{
			GroupID:        d.Get("group_id").(string),
			AccessPolicies: accessPolicies,
		}
		accessPolicyParams := identityClient.UpsertAccessPolicyRequest{
			RealmName:           d.Get("realm_name").(string),
			GroupAccessPolicies: groupAccessPolicies,
		}

		// Nothing to do in the case of a successful access policy upsert. It simply returns the group's access policies.
		_, err = toznySDK.UpsertAccessPolicies(ctx, accessPolicyParams)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func attributesFromState(d *schema.ResourceData) map[string][]string {
	attributes := map[string][]string{}
	if dAttributes, ok := d.GetOk("attribute"); ok {
		interfaceList := dAttributes.([]interface{})
		for _, attrMap := range interfaceList {
			pair := attrMap.(map[string]interface{})
			key := pair["key"].(string)
			values := pair["values"].([]interface{})
			if len(values) > 0 {
				attributes[key] = []string{}
				for _, value := range values {
					attributes[key] = append(attributes[key], value.(string))
				}
			}
		}
	}
	return attributes
}

func attributesToState(attributes map[string][]string) []interface{} {
	var stateAttributes []interface{}
	for key, values := range attributes {
		attrMap := map[string]interface{}{}
		attrMap["key"] = key
		attrMap["values"] = values
		stateAttributes = append(stateAttributes, attrMap)
	}
	return stateAttributes
}
