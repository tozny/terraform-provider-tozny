package tozny

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmProviderMapper returns the schema and methods for provisioning a Tozny Realm Provider Mapper.
func resourceRealmProviderMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmProviderMapperCreate,
		ReadContext:   resourceRealmProviderMapperRead,
		DeleteContext: resourceRealmProviderMapperDelete,
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
			"provider_id": {
				Description: "Service defined unique identifier for the provider to associate the mapper with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the realm to associate the provider with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"provider_mapper_id": {
				Description: "Service defined unique identifier for the provider mapper.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "User defined name for the provider mapper.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"provider_type": {
				Description: "The type of the provider mapper. Valid values are `msad-user-account-control-mapper`, `msad-lds-user-account-control-mapper`, `group-ldap-mapper`, `user-attribute-ldap-mapper`, `role-ldap-mapper`, `hardcoded-ldap-role-mapper`, `full-name-ldap-mapper`, `hardcoded-ldap-group-mapper`, `hardcoded-ldap-attribute-mapper`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"groups_dn": {
				Description: "LDAP DN where are groups of this tree saved. For example 'ou=groups,dc=example,dc=org'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_name_attribute": {
				Description: "Name of LDAP attribute, which is used in group objects for name and RDN of group. Usually it will be 'cn' . In this case typical group/role object may have DN like 'cn=Group1,ou=groups,dc=example,dc=org'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_object_classes": {
				Description: "Object class (or classes) of the group object. In typical LDAP deployment it could be 'groupOfNames' . In Active Directory it's usually 'group'.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"preserve_group_inheritance": {
				Description: "Flag whether group inheritance from LDAP should be propagated to the Realm. If false, then all LDAP groups will be mapped as flat top-level groups in the Realm. Otherwise group inheritance is preserved into the Realm, but the group sync might fail if LDAP structure contains recursions or multiple parent groups per child groups.",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
			"ignore_missing_groups": {
				Description: "Whether missing groups in the hierarchy should be ignored.",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
			"member_of_attribute": {
				Description: "Used just when `identity_groups_retrieval_strategy` is GET_GROUPS_FROM_USER_MEMBEROF_ATTRIBUTE . It specifies the name of the LDAP attribute on the LDAP identity, which contains the groups, which the identity is a member of. Usually it will be 'memberOf'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"membership_attribute": {
				Description: "Name of LDAP attribute on group, which is used for membership mappings. Usually it will be 'member' .However when 'Membership Attribute Type' is 'UID' then 'Membership LDAP Attribute' could be typically 'memberUid'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"membership_attribute_type": {
				Description: "DN means that LDAP group has it's members declared in form of their full DN. For example 'member: uid=john,ou=users,dc=example,dc=com' . UID means that LDAP group has it's members declared in form of pure user uids. For example 'memberUid: john'. Valid values are `DN` or `UID`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"membership_identity_attribute": {
				Description: "Used just if Membership Attribute Type is UID. It is name of LDAP attribute on user, which is used for membership mappings. Usually it will be 'uid' . For example if value of 'Membership User LDAP Attribute' is 'uid' and LDAP group has 'memberUid: john', then it is expected that particular LDAP user will have attribute 'uid: john'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"mode": {
				Description: "LDAP_ONLY means that all group mappings of users are retrieved from LDAP and saved into LDAP. READ_ONLY is Read-only LDAP mode where group mappings are retrieved from both LDAP and DB and merged together. New group joins are not saved to LDAP but to the Realm. IMPORT is Read-only LDAP mode where group mappings are retrieved from LDAP just at the time when user is imported from LDAP and then they are saved to the realm.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"identity_groups_retrieval_strategy": {
				Description: "Specify how to retrieve groups of user. LOAD_GROUPS_BY_MEMBER_ATTRIBUTE means that roles of user will be retrieved by sending LDAP query to retrieve all groups where 'member' is our user. GET_GROUPS_FROM_USER_MEMBEROF_ATTRIBUTE means that groups of user will be retrieved from 'memberOf' attribute of our user. Or from the other attribute specified by 'Member-Of LDAP Attribute' . LOAD_GROUPS_BY_MEMBER_ATTRIBUTE_RECURSIVELY is applicable just in Active Directory and it means that groups of user will be retrieved recursively with usage of LDAP_MATCHING_RULE_IN_CHAIN Ldap extension.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"drop_missing_groups_on_sync": {
				Description: "If this flag is true, then during sync of groups from LDAP to the Realm, we will keep just those Realm groups, which still exists in LDAP. Rest will be deleted.",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func fetchGroupObjectClasses(terraformData *schema.ResourceData) []string {
	terraformGroupObjectClasses := terraformData.Get("group_object_classes").([]interface{})

	groupObjectClasses := []string{}

	for _, terraformGroupObjectClass := range terraformGroupObjectClasses {
		groupObjectClasses = append(groupObjectClasses, terraformGroupObjectClass.(string))
	}

	return groupObjectClasses
}

func resourceRealmProviderMapperCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	createProviderRequest := identityClient.CreateRealmProviderMapperRequest{
		RealmName:  d.Get("realm_name").(string),
		ProviderID: d.Get("provider_id").(string),
		ProviderMapper: identityClient.ProviderMapper{
			Type:                            d.Get("provider_type").(string),
			Name:                            d.Get("name").(string),
			GroupsDN:                        d.Get("groups_dn").(string),
			GroupNameAttribute:              d.Get("group_name_attribute").(string),
			GroupObjectClasses:              fetchGroupObjectClasses(d),
			PreserveGroupInheritance:        d.Get("preserve_group_inheritance").(bool),
			IgnoreMissingGroups:             d.Get("ignore_missing_groups").(bool),
			MemberOfAttribute:               d.Get("member_of_attribute").(string),
			MembershipAttribute:             d.Get("membership_attribute").(string),
			MembershipAttributeType:         d.Get("membership_attribute_type").(string),
			Mode:                            d.Get("mode").(string),
			MembershipIdentityAttribute:     d.Get("membership_identity_attribute").(string),
			IdentityGroupsRetrievalStrategy: d.Get("identity_groups_retrieval_strategy").(string),
			DropMissingGroupsOnSync:         d.Get("drop_missing_groups_on_sync").(bool),
		},
	}

	providerMapper, err := toznySDK.CreateRealmProviderMapper(ctx, createProviderRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	providerMapperID := providerMapper.ID

	d.Set("provider_mapper_id", providerMapperID)

	// Associate created Realm Provider Mapper with Terraform state and signal success
	d.SetId(providerMapperID)

	return diags
}

func resourceRealmProviderMapperRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	describeProviderMapperRequest := identityClient.DescribeRealmProviderMapperRequest{
		RealmName:        d.Get("realm_name").(string),
		ProviderID:       d.Get("provider_id").(string),
		ProviderMapperID: d.Get("provider_mapper_id").(string),
	}

	providerMapper, err := toznySDK.DescribeRealmProviderMapper(ctx, describeProviderMapperRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	// to suppress spurious diffs only update the group_object_classes if
	// there is an upstream change
	configuredGroupObjectClasses := fetchGroupObjectClasses(d)

	actualGroupObjectClasses := providerMapper.GroupObjectClasses

	if len(configuredGroupObjectClasses) != len(actualGroupObjectClasses) {
		configuredGroupObjectClasses = actualGroupObjectClasses
	}

	for _, configuredGroupObjectClass := range configuredGroupObjectClasses {
		var match bool
		for _, actualGroupObjectClass := range actualGroupObjectClasses {
			if actualGroupObjectClass == configuredGroupObjectClass {
				match = true
				break
			}
		}
		if !match {
			configuredGroupObjectClasses = actualGroupObjectClasses
			break
		}
	}

	d.Set("provider_type", providerMapper.Type)
	d.Set("name", providerMapper.Name)
	d.Set("groups_dn", providerMapper.GroupsDN)
	d.Set("group_name_attribute", providerMapper.GroupNameAttribute)
	d.Set("group_object_classes", configuredGroupObjectClasses)
	d.Set("preserve_group_inheritance", providerMapper.PreserveGroupInheritance)
	d.Set("ignore_missing_groups", providerMapper.IgnoreMissingGroups)
	d.Set("member_of_attribute", providerMapper.MemberOfAttribute)
	d.Set("membership_attribute", providerMapper.MembershipAttribute)
	d.Set("membership_attribute_type", providerMapper.MembershipAttributeType)
	d.Set("mode", providerMapper.Mode)
	d.Set("membership_identity_attribute", providerMapper.MembershipIdentityAttribute)
	d.Set("identity_groups_retrieval_strategy", providerMapper.IdentityGroupsRetrievalStrategy)
	d.Set("drop_missing_groups_on_sync", providerMapper.DropMissingGroupsOnSync)

	return diags
}

func resourceRealmProviderMapperDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	toznySDK, err := MakeToznySDK(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	deleteRealmProviderMapperParams := identityClient.DeleteRealmProviderMapperRequest{
		RealmName:        d.Get("realm_name").(string),
		ProviderID:       d.Get("provider_id").(string),
		ProviderMapperID: d.Get("provider_mapper_id").(string),
	}

	err = toznySDK.DeleteRealmProviderMapper(ctx, deleteRealmProviderMapperParams)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
