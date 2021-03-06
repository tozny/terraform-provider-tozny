# tozny_account Resource

Resource for provisioning (Create only) a Tozny account, the primary top level resource for all other resources provided by Tozny (e.g. Clients, Realms, Application). This resource will provision a Tozny account, either from user provided credentials, or autogenerated credentials that are then persisted to disk.

## Example Usage

```hcl
# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_email_salt" {
  length = 8
  special = false
}

# A resource for provisioning a Tozny account from explicit configuration
resource "tozny_account" "pregenerated_tozny_account_credentials" {
  client_credentials_save_filepath = "./tozny_client_credentials.json"
  profile {
    name = "TozFormed"
    email = "terraform+${random_string.account_email_salt.result}@tozny.com"
    authentication_salt = "xXp-SXoBEmFA5aU2k50-4g"
    signing_key {
      ed25519_public_key = "GxESnvi-4wMdH4IKP1cHsI7IfthpzhF2FRVrn0Z3RBY"
    }
    encoding_salt = "A8E2O5rzkrL5R3HeV7py6A"
    paper_authentication_salt = "7FspmhLxmZ15xi_ZJs-e3g"
    paper_encoding_salt = "0SLzQBQD611ihctI85PdqA"
    paper_signing_key {
      ed25519_public_key = "HLJ_kTSnKFXfCnh-y0c83Uqq1JVk6RpPRgd-zPaSl6Y"
    }
  }
  account {
    company = "Terrafirma"
    plan = "free0"
    public_key {
      ed25519_public_key = "WxqO3mYCwSMPH99I2PJpUI_f6CiTvnoEfcjopyXGUnw"
    }
    signing_key {
       ed25519_public_key = "CjjbYPY21YGBsDaXRJ-hqv90ohT3kxckRsssXTow9rc"
    }
  }
}
```

```hcl
# A resource for provisioning a Tozny account from explicit configuration read from a local file.
# Tied for 1st most secure method in that no private information has to be passed to Terraform
resource "tozny_account" "pregenerated_tozny_account_credentials_from_file" {
  autogenerate_account_credentials = false
  account_credentials_filepath = "account.json"
  client_credentials_save_filepath = "./tozny_client_credentials.json"
}
```

```hcl
# A resource for provisioning a Tozny account from explicit configuration
resource "tozny_account" "pregenerated_tozny_account_credentials" {
  client_credentials_save_filepath = "./tozny_client_credentials.json"
  profile {
    name = "TozFormed"
    email = "terraform+${random_string.account_email_salt.result}@tozny.com"
    authentication_salt = "xXp-SXoBEmFA5aU2k50-4g"
    signing_key {
      ed25519_public_key = "GxESnvi-4wMdH4IKP1cHsI7IfthpzhF2FRVrn0Z3RBY"
    }
    encoding_salt = "A8E2O5rzkrL5R3HeV7py6A"
    paper_authentication_salt = "7FspmhLxmZ15xi_ZJs-e3g"
    paper_encoding_salt = "0SLzQBQD611ihctI85PdqA"
    paper_signing_key {
      ed25519_public_key = "HLJ_kTSnKFXfCnh-y0c83Uqq1JVk6RpPRgd-zPaSl6Y"
    }
  }
  account {
    company = "Terrafirma"
    plan = "free0"
    public_key {
      ed25519_public_key = "WxqO3mYCwSMPH99I2PJpUI_f6CiTvnoEfcjopyXGUnw"
    }
    signing_key {
       ed25519_public_key = "CjjbYPY21YGBsDaXRJ-hqv90ohT3kxckRsssXTow9rc"
    }
  }
   persist_credentials_to = "file"
}
```

```hcl
# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to user specified filepath for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  client_credentials_save_filepath = "./tozny_client_credentials.json"
}
```

```hcl
# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to terraform state for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account_terraform" {
  autogenerate_account_credentials = true
  persist_credentials_to = "terraform"
}
```
## Argument Reference

### Top-Level Arguments

* `autogenerate_account_credentials` - (Optional) Whether Terraform should generate credentials for a provisioned account. Defaults to `false`.
* `persist_credentials_to` - (Optional) Where to persist the generated credentials. "none", "file", or "terraform". Default: none
* `account_credentials_filepath` - (Optional) The filepath where account credentials will be loaded from.
* `client_credentials_save_filepath` - (Optional) The filepath where client credentials will be persisted. Defaults to `tozny_client_credentials.json`
* `profile` - (Optional) The filepath where client credentials will be persisted. The account creator's profile settings.
* `account` - (Optional) Account wide settings.
* `config` - (Computed) A JSON representation of the generated credentials, only populated when `persist_credentials_to` is set to "terraform"

### Account Arguments

* `company` - (Optional) Billing name of the account holder's organization.
* `plan` - (Optional) Tozny Billing plan associated with the account.
* `public_key` - (Required) The public key of the keypair used for account level encryption operations.
* `signing_key` - (Required) The public key of the keypair used for account level signing operations.

### Profile Arguments

* `account_id` - (Computed) The unique server defined identifier for the account.
* `verified` - (Computed) Whether or not the email for the account profile has been verified.
* `name` - (Optional) User defined identifier for the account registration profile.
* `email` - (Required) The email for the account registration profile.
* `authentication_salt` - (Required) The salt used to generate the authentication keypair.
* `signing_key` - (Required) The public key generated using the authentication salt used to generate the encryption keypair.
* `encoding_salt` - (Required) The salt used to generate the encryption keypair.
* `paper_authentication_salt` - (Required) The salt used to generate the paper authentication keypair.
encryption keypair.
* `paper_encoding_salt` - (Required) The salt used to generate the paper encoding keypair.
* `paper_signing_key` - (Required) The paper public key generated using the authentication salt used to generate the encryption keypair.

### Client Public Key Schema

* `ed25519_public_key` - (Required) A public key from a keypair based off the Ed25519 curve.
* `p384_public_key` - (Required)  public key from a keypair based off the P384 curve.

### Encryption Public Key Schema

* `ed25519_public_key` - (Required) A public key from a keypair based off the Ed25519 curve.

## Attribute Reference

* `id` - Unique ID of the provisioned Account.
* `config` - A JSON representation of the generated credentials, only populated when `persist_credentials_to` is set to "terraform"
