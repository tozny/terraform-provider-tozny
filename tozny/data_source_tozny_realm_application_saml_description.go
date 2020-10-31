package tozny

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// dataSourceRealmApplicationSAMLDescription returns the schema and methods for retrieving the SAML XML description document for a Tozny Realm Application
func dataSourceRealmApplicationSAMLDescription() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRealmApplicationSAMLDescriptionRead,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when fetching this application's SAML description.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"realm_name": {
				Description: "The name of the realm the application is associated with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"application_id": {
				Description: "The application ID to retrieve the SAML description for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"format": {
				Description: "The format of the description to retrieve.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					identityClient.SAMLIdentityProviderDescriptionFormat,
					identityClient.SAMLKeycloakDescriptionFormat,
					identityClient.SAMLServiceProviderDescriptionFormat,
					identityClient.SAMLKeycloakSubsystemDescriptionFormat}, false),
			},
			"description": {
				Description: "SAML XML description document.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"description_save_filepath": {
				Description: "The filepath to save the SAML description to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func dataSourceRealmApplicationSAMLDescriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)
	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)
	if err != nil {
		return diag.FromErr(err)
	}

	fetchApplicationSAMLDescriptionParams := identityClient.FetchApplicationSAMLDescriptionRequest{
		RealmName:     strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID: d.Get("application_id").(string),
		Format:        d.Get("format").(string),
	}

	response, err := toznySDK.FetchApplicationSAMLDescription(ctx, fetchApplicationSAMLDescriptionParams)
	if err != nil {
		return diag.FromErr(err)
	}
	description := response.Description

	d.Set("description", description)

	fileSavePath := d.Get("description_save_filepath").(string)
	// If the file doesn't exist, create it, or overwrite the file if it exists
	f, err := os.OpenFile(fileSavePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return diag.FromErr(err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(description)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uuid.New().String())

	return diags
}
