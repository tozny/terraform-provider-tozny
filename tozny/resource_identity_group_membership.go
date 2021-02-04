package tozny

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceIdentityGroupMembership returns the schema and methods for configuring Tozny Realm default groups
func resourceIdentityGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityGroupMembershipCreateOrUpdate,
		ReadContext:   resourceIdentityGroupMembershipRead,
		DeleteContext: resourceIdentityGroupMembershipDelete,
		UpdateContext: resourceIdentityGroupMembershipCreateOrUpdate,
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
			},
			"identity_id": {
				Description: "The Tozny ID (Client ID) of the identity to map to join with the groups in group_ids",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_ids": {
				Description: "The IDs of the groups to make default for all users in the realm",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceIdentityGroupMembershipCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := strings.ToLower(d.Get("realm_name").(string))
	identityID := strings.ToLower(d.Get("identity_id").(string))
	groupList := schemaToStringSlice(d.Get("group_ids").([]interface{}))
	err = toznySDK.UpdateGroupMembership(ctx, identityClient.UpdateIdentityGroupMembershipRequest{
		RealmName:  realmName,
		IdentityID: identityID,
		Groups:     groupList,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	// The default groups live on even when empty
	// but we still need unique ID for Terraform's idempotent satisfaction
	d.SetId(uuid.New().String())

	return diags
}

func resourceIdentityGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := strings.ToLower(d.Get("realm_name").(string))
	identityID := strings.ToLower(d.Get("identity_id").(string))
	storedGroupList, _ := d.GetChange("group_ids")
	groupList := schemaToStringSlice(storedGroupList.([]interface{}))
	serverGroups, err := toznySDK.GroupMembership(ctx, identityClient.RealmIdentityRequest{
		RealmName:  realmName,
		IdentityID: identityID,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	// Use a map-set to read through groups in state-order using false to indicated
	// they have not been seen yet from the server state
	groupMapSet := map[string]bool{}
	for _, groupID := range groupList {
		groupMapSet[groupID] = false
	}
	// Update the map set with settings read from the server
	for _, group := range serverGroups.Groups {
		groupMapSet[group.ID] = true
	}

	// translate group list back into terraform state format
	updatedIDs := []interface{}{}
	for groupID, found := range groupMapSet {
		// found indicates this was seen in the server fetched data
		// otherwise it was only found in terraform state and is missing
		if found {
			updatedIDs = append(updatedIDs, groupID)
		}
	}

	d.Set("group_ids", updatedIDs)

	return diags
}

func resourceIdentityGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := strings.ToLower(d.Get("realm_name").(string))
	identityID := strings.ToLower(d.Get("identity_id").(string))

	err = toznySDK.UpdateGroupMembership(ctx, identityClient.UpdateIdentityGroupMembershipRequest{
		RealmName:  realmName,
		IdentityID: identityID,
		Groups:     []string{}, // replace with an empty slice to remove all current default groups
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
