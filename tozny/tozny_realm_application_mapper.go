package tozny

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmApplicationMapper returns the schema and methods for provisioning a Tozny Realm Application Mapper
func resourceRealmApplicationMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmApplicationMapperCreate,
		ReadContext:   resourceRealmApplicationMapperRead,
		DeleteContext: resourceRealmApplicationMapperDelete,
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
				Description: "The name of the Realm to provision the Application Mapper in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_id": {
				Description: "ID of the Application the Mapper is associated with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "User defined name for the application mapper.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"protocol": {
				Description:  "The identity protocol that this mapper will be applied to flows of. Valid values are `openid-connect`, `saml`.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{identityClient.ProtocolSAML, identityClient.ProtocolOIDC}, false),
			},
			"mapper_type": {
				Description: "The category of data this mapper is applied to. Valid values are `oidc-user-session-note-mapper`, `oidc-user-attribute-mapper`, `oidc-group-membership-mapper`, `saml-role-list-mapper`, `saml-user-property-mapper`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					identityClient.UserSessionNoteOIDCApplicationMapperType,
					identityClient.UserAttributeOIDCApplicationMapperType,
					identityClient.RoleListSAMLApplicationMapperType,
					identityClient.UserPropertySAMLApplicationMapperType,
					identityClient.GroupMembershipOIDCApplicationMapperType,
				}, false),
			},
			"user_session_note": {
				Description: "Name of stored user session note within the UserSessionModel.note map.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"user_attribute": {
				Description: "Name of stored user attribute within the UserModel.attribute map.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"token_claim_name": {
				Description: "Name of the claim to insert into the token. This can be a fully qualified name like 'address.street'. In this case, a nested json object will be created. To prevent nesting and use dot literally, escape the dot with backslash (\\.)",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"claim_json_type": {
				Description:  "JSON type that should be used to populate the json claim in the token. Valid `String`, `int`, `bool`, `long`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{identityClient.ClaimJSONStringType, identityClient.ClaimJSONLongType, identityClient.ClaimJSONIntType, identityClient.ClaimJSONBooleanType}, false),
			},
			"add_to_id_token": {
				Description: "Indicates if the claim should be added to the id token. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"add_to_access_token": {
				Description: "Indicates if the claim should be added to the access token. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"add_to_user_info": {
				Description: "Indicates if the claim should be added to the user info. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"multivalued": {
				Description: "Indicates if attribute supports multiple values. If true, then the list of all values of this attribute will be set as claim. If false, then just first value will be set as claim.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"aggregate_attribute_values": {
				Description: "Indicates if attribute values should be aggregated with the group attributes. If using OpenID Connect mapper the multivalued option needs to be enabled too in order to get all the values. Duplicated values are discarded and the order of values is not guaranteed with this option.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"saml_attribute_name": {
				Description: "Name of the SAML attribute that should be used for mapping an identities name.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"saml_attribute_name_format": {
				Description:  "Format to use for the name attribute for the SAML protocol, valid values are `Basic` `URI Reference` or `Unspecified`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{identityClient.BasicSAMLAttributeNameFormat, identityClient.UnspecifiedSAMLAttributeNameFormat, identityClient.URIReferenceSAMLAttributeNameFormat}, false),
			},
			"friendly_name": {
				Description: "Standard SAML attribute setting. An optional, more human-readable form of the attribute's name that can be provided if the actual attribute name is cryptic.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"role_attribute_name": {
				Description: "Name of the SAML attribute you want to put your roles into. i.e. 'Role', 'memberOf'.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"property": {
				Description: "Name of the property method in the UserModel interface. For example, a value of 'email' would reference the UserModel.getEmail() method.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"single_role_attribute": {
				Description: "If true, all roles will be stored under one attribute with multiple attribute values.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"application_mapper_id": {
				Description: "Server defined unique identifier for the Application Mapper.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"full_group_path": {
				Description: "If true, will include the full group path in tokens when the group-mapper is created.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
		},
	}
}

func resourceRealmApplicationMapperCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createApplicationMapperParams := identityClient.CreateRealmApplicationMapperRequest{
		RealmName:     strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID: d.Get("application_id").(string),
		ApplicationMapper: identityClient.ApplicationMapper{
			Name:                     d.Get("name").(string),
			Protocol:                 d.Get("protocol").(string),
			MapperType:               d.Get("mapper_type").(string),
			UserSessionNote:          d.Get("user_session_note").(string),
			UserAttribute:            d.Get("user_attribute").(string),
			FullPath:                 d.Get("full_group_path").(bool),
			TokenClaimName:           d.Get("token_claim_name").(string),
			ClaimJSONType:            d.Get("claim_json_type").(string),
			AddToIDToken:             d.Get("add_to_id_token").(bool),
			AddToAccessToken:         d.Get("add_to_access_token").(bool),
			AddToUserInfo:            d.Get("add_to_user_info").(bool),
			Multivalued:              d.Get("multivalued").(bool),
			AggregateAttributeValues: d.Get("aggregate_attribute_values").(bool),
			RoleAttributeName:        d.Get("role_attribute_name").(string),
			Property:                 d.Get("property").(string),
			FriendlyName:             d.Get("friendly_name").(string),
			SAMLAttributeName:        d.Get("saml_attribute_name").(string),
			SAMLAttributeNameFormat:  d.Get("saml_attribute_name_format").(string),
			SingleRoleAttribute:      d.Get("single_role_attribute").(bool),
		},
	}
	applicationMapper, err := toznySDK.CreateRealmApplicationMapper(ctx, createApplicationMapperParams)
	if err != nil {
		return diag.FromErr(err)
	}

	applicationMapperID := applicationMapper.ID
	d.Set("application_mapper_id", applicationMapperID)
	d.SetId(applicationMapperID)

	return diags
}

func resourceRealmApplicationMapperRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	applicationMapper, err := toznySDK.DescribeRealmApplicationMapper(ctx, identityClient.DescribeRealmApplicationMapperRequest{
		RealmName:           strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID:       d.Get("application_id").(string),
		ApplicationMapperID: d.Get("application_mapper_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", applicationMapper.Name)
	d.Set("protocol", applicationMapper.Protocol)
	d.Set("mapper_type", applicationMapper.MapperType)
	d.Set("user_session_note", applicationMapper.UserSessionNote)
	d.Set("user_attribute", applicationMapper.UserAttribute)
	d.Set("full_group_path", applicationMapper.FullPath)
	d.Set("token_claim_name", applicationMapper.TokenClaimName)
	d.Set("claim_json_type", applicationMapper.ClaimJSONType)
	d.Set("add_to_id_token", applicationMapper.AddToIDToken)
	d.Set("add_to_access_token", applicationMapper.AddToAccessToken)
	d.Set("add_to_user_info", applicationMapper.AddToUserInfo)
	d.Set("multivalued", applicationMapper.Multivalued)
	d.Set("aggregate_attribute_values", applicationMapper.AggregateAttributeValues)
	d.Set("role_attribute_name", applicationMapper.RoleAttributeName)
	d.Set("property", applicationMapper.Property)
	d.Set("friendly_name", applicationMapper.FriendlyName)
	d.Set("saml_attribute_name", applicationMapper.SAMLAttributeName)
	d.Set("saml_attribute_name_format", applicationMapper.SAMLAttributeNameFormat)
	d.Set("single_role_attribute", applicationMapper.SingleRoleAttribute)

	return diags
}

func resourceRealmApplicationMapperDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	err = toznySDK.DeleteRealmApplicationMapper(ctx, identityClient.DeleteRealmApplicationMapperRequest{
		RealmName:           strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID:       d.Get("application_id").(string),
		ApplicationMapperID: d.Get("application_mapper_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")

	return diags
}
