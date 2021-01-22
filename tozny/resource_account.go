package tozny

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tozny/e3db-clients-go/accountClient"
	"github.com/tozny/e3db-go/v2"
)

// AccountCredentialsFile wraps partial or complete
// account credentials suitable for
// serializing to and from JSON for SDK consumption
type AccountCredentialsFile struct {
	APIEndpoint     string `json:"api_url"`
	AccountUsername string `json:"account_username"`
	AccountPassword string `json:"account_password"`
	Account         accountClient.Account
	Profile         accountClient.Profile
}

// EncryptionPublicKeySchema returns the Terraform schema for
// describing the Tozny Public Key of the keypair used for encryption operations.
func EncryptionPublicKeySchema() *schema.Resource {
	scheme := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ed25519_public_key": {
				Description: "A public key from a keypair based off the Ed25519 curve.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
	return scheme
}

// ClientPublicKeySchema returns the Terraform schema for
// describing the Public Key(s) of the keypair used for Tozny client level operations.
func ClientPublicKeySchema() *schema.Resource {
	scheme := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ed25519_public_key": {
				Description: "A public key from a keypair based off the Ed25519 curve.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"p384_public_key": {
				Description: "A public key from a keypair based off the P384 curve.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
	return scheme
}

// ClientSchema returns the Terraform schema for
// describing a Tozny client.
func ClientSchema() *schema.Resource {
	scheme := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Description: "The server defined unique identifier for a Tozny cryptography client.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "User defined identifier for the client.",
				Type:        schema.TypeString,
				Required:    false,
				Computed:    true,
				ForceNew:    true,
			},
			"public_key": {
				Description: "The public key of the keypair used for client level encryption operations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				ForceNew:    true,
				Required:    true,
				Elem:        ClientPublicKeySchema(),
			},
			"signing_key": {
				Description: "The public key of the keypair used for account level signing operations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				ForceNew:    true,
				Required:    true,
				Elem:        EncryptionPublicKeySchema(),
			},
			"api_key_id": {
				Description: "Public API credential for authenticating requests.",
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    false,
				Computed:    true,
				ForceNew:    true,
			},
			"api_secret_key": {
				Description: "Private API credential for authenticating requests.",
				Type:        schema.TypeString,
				Required:    false,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"enabled": {
				Description: "Whether or not the client is enabled for account & cryptographic operations",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
	return scheme
}

// resourceAccount returns the schema and methods for provisioning a Tozny account.
func resourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountCreate,
		ReadContext:   resourceAccountRead,
		DeleteContext: resourceAccountDelete,
		Schema: map[string]*schema.Schema{
			"persist_credentials_to": {
				Description:  "Where to persist the generated account credentials. Default: none, they are not persisted.",
				Type:         schema.TypeString,
				Default:      "none",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"none", "file", "terraform"}, false),
			},
			"autogenerate_account_credentials": {
				Description:   "Whether Terraform should generate credentials for a provisioned account.",
				Type:          schema.TypeBool,
				Default:       false,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"account", "profile"},
			},
			"account_credentials_filepath": {
				Description: "The filepath where account credentials will be loaded from.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"client_credentials_save_filepath": {
				Description: "The filepath where client credentials will be persisted.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "tozny_client_credentials.json",
				ForceNew:    true,
			},
			"config": {
				Description: "The client configuration as a JSON string. Only populated when persist_credentails_to is set to 'terraform'",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"profile": {
				Description: "The account creator's profile settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Description: "The unique server defined identifier for the account.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "User defined identifier for the account registration profile.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"email": {
							Description: "The email for the account registration profile.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"authentication_salt": {
							Description: "The salt used to generate the authentication keypair.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"signing_key": {
							Description: "The public key generated using the authentication salt used to generate the encryption keypair.",
							Type:        schema.TypeList,
							MaxItems:    1,
							MinItems:    1,
							ForceNew:    true,
							Required:    true,
							Elem:        EncryptionPublicKeySchema(),
						},
						"encoding_salt": {
							Description: "The salt used to generate the encryption keypair.",
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
						},
						"paper_authentication_salt": {
							Description: "The salt used to generate the paper authentication keypair.",
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
						},
						"paper_encoding_salt": {
							Description: "The salt used to generate the paper encoding keypair.",
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
						},
						"paper_signing_key": {
							Description: "The paper public key generated using the authentication salt used to generate the encryption keypair.",
							Type:        schema.TypeList,
							MaxItems:    1,
							MinItems:    1,
							ForceNew:    true,
							Required:    true,
							Elem:        EncryptionPublicKeySchema(),
						},
						"verified": {
							Description: "Whether or not the email for the account profile has been verified.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
			"account": {
				Description: "Account wide settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"company": {
							Description: "Billing name of the account holder's organization.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"plan": {
							Description: "Tozny Billing plan associated with the account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"public_key": {
							Description: "The public key of the keypair used for account level encryption operations.",
							Type:        schema.TypeList,
							MaxItems:    1,
							MinItems:    1,
							ForceNew:    true,
							Required:    true,
							Elem:        ClientPublicKeySchema(),
						},
						"signing_key": {
							Description: "The public key of the keypair used for account level signing operations.",
							Type:        schema.TypeList,
							MaxItems:    1,
							MinItems:    1,
							ForceNew:    true,
							Required:    true,
							Elem:        EncryptionPublicKeySchema(),
						},
					},
				},
			},
		},
	}
}

/*
  resourceAccountCreate creates an account based off (file driven or Terraform
  schema derived) + provider configuration using the following algorithm
   if auto-generate
     if account credentials file
       error - conflict
     else
       if no username on provider
         error - missing data
       if no password on provider
         auto-generate password
       use username & password to derive account credentials
       create account
       save client & account config to file
   else if NOT auto generate
    if account credentials file
      load credentials
      create account
      save client config to file
    else
      parse config from Terraform
      create account
      save client config to file
*/
func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK := m.(TerraformToznySDKResult).SDK

	var createAccountParams accountClient.CreateAccountRequest

	autoGenerateKey := "autogenerate_account_credentials"

	persistKey := "persist_credentials_to"
	persistTo := d.Get(persistKey).(string)

	credentialsFilepathKey := "account_credentials_filepath"
	accountCredentialsFilepath := d.Get(credentialsFilepathKey).(string)

	saveFilepathKey := "client_credentials_save_filepath"
	saveFilepath := d.Get(saveFilepathKey).(string)

	var accountUsername, accountPassword string

	apiEndpoint := toznySDK.APIEndpoint

	var sdkV3Config e3db.ToznySDKJSONConfig

	var accountID string

	if d.Get(autoGenerateKey).(bool) {
		if accountCredentialsFilepath != "" {
			return diag.Errorf("Only one of %s or %s can be specified", autoGenerateKey, credentialsFilepathKey)
		}

		if persistTo != "file" && persistTo != "terraform" {
			return diag.Errorf("Can not auto-generate credentials if no persistance is defined in %q", persistKey)
		} else if persistTo == "file" && saveFilepath == "" {
			return diag.Errorf("%s must be supplied if %s is set to %q", saveFilepathKey, persistKey, "file")
		}

		accountUsername, accountPassword = toznySDK.AccountUsername, toznySDK.AccountPassword

		if accountUsername == "" {
			return diag.Errorf("Must specify %q with provider config when auto generating account resource.", "account_username")
		}

		if accountPassword == "" {
			accountPassword = uuid.New().String()
		}

		createdAccount, err := toznySDK.Register(ctx, accountUsername, accountUsername, accountPassword)

		if err != nil {
			return diag.FromErr(err)
		}

		sdkV3Config = e3db.ToznySDKJSONConfig{
			ConfigFile: e3db.ConfigFile{
				Version:     createdAccount.Account.Config.Version,
				APIBaseURL:  createdAccount.Account.Config.APIURL,
				APIKeyID:    createdAccount.Account.Config.APIKeyID,
				APISecret:   createdAccount.Account.Config.APISecret,
				ClientID:    createdAccount.Account.Config.ClientID,
				ClientEmail: createdAccount.Account.Config.ClientEmail,
				PublicKey:   createdAccount.Account.Config.PublicKey,
				PrivateKey:  createdAccount.Account.Config.PrivateKey,
			},
			AccountPassword:   accountPassword,
			AccountUsername:   strings.ToLower(accountUsername),
			PublicSigningKey:  createdAccount.Account.Config.PublicSigningKey,
			PrivateSigningKey: createdAccount.Account.Config.PrivateSigningKey,
		}

		accountID = createdAccount.Account.AccountID

	} else {
		if accountCredentialsFilepath != "" {
			var accountCredentials AccountCredentialsFile

			bytes, err := ioutil.ReadFile(accountCredentialsFilepath)

			if err != nil {
				return diag.FromErr(err)
			}

			err = json.Unmarshal(bytes, &accountCredentials)

			if err != nil {
				return diag.FromErr(err)
			}

			createAccountParams = accountClient.CreateAccountRequest{
				Profile: accountCredentials.Profile,
				Account: accountCredentials.Account,
			}
		} else {
			profile := d.Get("profile").([]interface{})[0].(map[string]interface{})

			profileSigningKey := profile["signing_key"].([]interface{})[0].(map[string]interface{})

			profilePaperSigningKey := profile["paper_signing_key"].([]interface{})[0].(map[string]interface{})

			account := d.Get("account").([]interface{})[0].(map[string]interface{})

			accountPublicKey := account["public_key"].([]interface{})[0].(map[string]interface{})

			accountSigningKey := account["signing_key"].([]interface{})[0].(map[string]interface{})

			createAccountParams = accountClient.CreateAccountRequest{
				Profile: accountClient.Profile{
					Name:               profile["name"].(string),
					Email:              profile["email"].(string),
					AuthenticationSalt: profile["authentication_salt"].(string),
					EncodingSalt:       profile["encoding_salt"].(string),
					SigningKey: accountClient.EncryptionKey{
						Ed25519: profileSigningKey["ed25519_public_key"].(string),
					},
					PaperAuthenticationSalt: profile["paper_authentication_salt"].(string),
					PaperEncodingSalt:       profile["paper_encoding_salt"].(string),
					PaperSigningKey: accountClient.EncryptionKey{
						Ed25519: profilePaperSigningKey["ed25519_public_key"].(string),
					},
				},
				Account: accountClient.Account{
					Company: account["company"].(string),
					Plan:    account["plan"].(string),
					PublicKey: accountClient.ClientKey{
						Curve25519: accountPublicKey["ed25519_public_key"].(string),
					},
					SigningKey: accountClient.EncryptionKey{
						Ed25519: accountSigningKey["ed25519_public_key"].(string),
					},
				},
			}
		}

		// create account
		createAccountResponse, err := toznySDK.CreateAccount(ctx, createAccountParams)

		if err != nil {
			return diag.FromErr(err)
		}

		accountID = createAccountResponse.Profile.AccountID

		// save client config to file
		sdkV3Config = e3db.ToznySDKJSONConfig{
			ConfigFile: e3db.ConfigFile{
				Version:     2,
				APIBaseURL:  apiEndpoint,
				APIKeyID:    createAccountResponse.Account.Client.APIKeyID,
				APISecret:   createAccountResponse.Account.Client.APISecretKey,
				ClientID:    createAccountResponse.Account.Client.ClientID,
				ClientEmail: createAccountResponse.Profile.Email,
				PublicKey:   createAccountResponse.Account.Client.PublicKey.Curve25519,
				PrivateKey:  "",
			},
			AccountPassword:   accountPassword,
			AccountUsername:   strings.ToLower(accountUsername),
			PublicSigningKey:  createAccountResponse.Account.Client.SigningKey.Ed25519,
			PrivateSigningKey: "",
		}
	}
	clientCredentialsJSONBytes, err := json.Marshal(sdkV3Config)

	if err != nil {
		return diag.FromErr(err)
	}

	switch persistTo {
	case "file":
		err = ioutil.WriteFile(d.Get("client_credentials_save_filepath").(string), clientCredentialsJSONBytes, 0644)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("config", "")
		break
	case "terraform":
		d.Set("config", string(clientCredentialsJSONBytes))
		break
	default:
		d.Set("config", "")
		break
	}

	// Associate created account with Terraform state and signal success
	d.SetId(accountID)

	return diags
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// No op as there is no directly corresponding "Read" operation" for an account
	// Could be be updated in the future to read the current billing status for an account
	return diags
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// While 'real Gs' / Tozny accounts never die nor can they be deleted from the API,
	// at least we can remove them from Terraform state ;-)
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}
