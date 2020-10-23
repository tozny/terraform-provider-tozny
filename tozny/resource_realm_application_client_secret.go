package tozny

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourceRealmApplicationClientSecret returns the schema and methods for provisioning a Tozny Realm Application Secret
func resourceRealmApplicationClientSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmApplicationClientSecretRead,
		ReadContext:   resourceRealmApplicationClientSecretRead,
		DeleteContext: resourceRealmApplicationClientSecretDelete,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description: "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this application client secret.",
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
				Description: "The application ID to retrieve the client secret for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"secret": {
				Description: "OIDC Client secret for the application. Will always be empty if `persist_client_secret_to_terraform` is `false`.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"persist_client_secret_to_terraform": {
				Description: "Whether or not the client secret should be persisted to terraform. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
			},
			"client_secret_save_filepath": {
				Description: "The filepath to save the client secret to. If not specified the secret will not be saved to the filesystem.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRealmApplicationClientSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)
	toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)
	if err != nil {
		return diag.FromErr(err)
	}

	fetchApplicationClientSecretParams := identityClient.FetchApplicationSecretRequest{
		RealmName:     strings.ToLower(d.Get("realm_name").(string)),
		ApplicationID: d.Get("application_id").(string),
	}

	applicationSecret, err := toznySDK.FetchApplicationSecret(ctx, fetchApplicationClientSecretParams)
	if err != nil {
		return diag.FromErr(err)
	}
	secret := applicationSecret.Secret

	saveToState := d.Get("persist_client_secret_to_terraform").(bool)
	if saveToState {
		d.Set("secret", secret)
	}

	fileSavePath := d.Get("client_secret_save_filepath").(string)
	if fileSavePath != "" {
		// If the file doesn't exist, create it, or overwrite the file if it exists
		f, err := os.OpenFile(fileSavePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return diag.FromErr(err)
		}
		defer f.Close()
		if _, err := f.Write([]byte(secret)); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(uuid.New().String())

	return diags
}

func resourceRealmApplicationClientSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Nothing to delete as the secret is tied to the application
	d.SetId("")
	return diags
}
