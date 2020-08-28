# tozny_realm Resource

Resource for provisioning a TozID Realm, the primary top level resource for all other resources provided by TozID (e.g. Identities, Groups, Roles & Applications).

## Example Usage

```hcl
# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint = "https://dev.e3db.com"
  account_username = "test+${random_string.account_username_salt.result}@tozny.com"
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_username_salt" {
  length = 8
  special = false
}

# Generate a random string for use in creating
# realms across environments or executions conflict free
resource "random_string" "random_realm_name" {
  length = 8
  special = false
}

# Local variables for defining where to store and find local Tozny client credentials
locals {
  tozny_client_credentials_filepath = "./tozny_client_credentials.json"
}

# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to user specified filepath for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  client_credentials_save_filepath = local.tozny_client_credentials_filepath
}

# A resource for provisioning a TozID Realm
# using local filebased credentials for a Tozny Client with permissions to manage Realms.
resource "tozny_realm" "my_organizations_realm" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name = random_string.random_realm_name.result
  sovereign_name = "Administrator"
}
```

## Argument Reference

### Top-Level Arguments

* `realm_id` - (Computed)Service defined unique identifier for the realm.
* `domain` - (Computed) Service defined & externally unique reference for the realm.
* `admin_url` - (Computed) URL for realm administration console.
* `active` - (Computed) Whether the realm is active for applications and identities to consume.
* `broker_identity_tozny_id` - (Computed) The Tozny Client ID associated with the Identity used to broker interactions between the realm and it's Identities. Will be empty if no realm broker Identity has been registered.
* `client_credentials_filepath` - (Optional) The filepath to Tozny client credentials for the provider to use when provisioning this realm.
* `realm_name` - (Required) User defined identifier for the realm.
* `sovereign_name` - (Required) User defined sovereign identifier.
* `sovereign` - (Computed) The admin identity for a realm.

### Sovereign Arguments

* `id` - (Computed) Service defined unique identifier for the sovereign.
* `name` - (Computed) User defined sovereign identifier.


## Attribute Reference

* `id` - Unique ID of the provisioned Account.