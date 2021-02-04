# tozny_realm_identity Resource

Resource to manage an identity for a TozID Realm. Useful for supporting service users.

This resource requires that the account username and password be supplied to the provider either via explicit provider settings or file based credentials.

This resource currently only supports supplying a password for the user. This is saved to terraform state for use in other modules. The intent is this resource creates service
users which are embedded in other providers and resources. At this time, it is not suggested to create identities for end users with this resource.

## Example Usage
```hcl
# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint = "http://platform.local.tozny.com:8000"
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

# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to user specified filepath for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  persist_credentials_to = "terraform"
}

# A resource for provisioning a TozID Realm
# using local file based credentials for a Tozny Client with permissions to manage Realms.
resource "tozny_realm" "my_organizations_realm" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = random_string.random_realm_name.result
  sovereign_name = "Administrator"
}

# A resource for provisioning a token that can be used to register Tozny clients
# such as a realm broker identity
resource "tozny_client_registration_token" "realm_registration_token" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "${tozny_realm.my_organizations_realm.realm_name}DefaultClientRegistrationToken"
  allowed_registration_client_types = ["general", "identity", "broker"]
  enabled = true
  one_time_use = false
}

# A resource for provisioning an identity that can be used to delegate authority for
# realm activities such as email recovery
resource "tozny_realm_broker_identity" "broker_identity" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
     tozny_client_registration_token.realm_registration_token
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name = tozny_realm.my_organizations_realm.realm_name
  name = "broker${tozny_realm.my_organizations_realm.realm_name}"
  persist_credentials_to = "terraform"
}

# A resource for delegating authority to another Tozny client
# (either hosted by Tozny or some other third party) to broker realm
# activities (such as email recovery).
resource "tozny_realm_broker_delegation" "allow_tozny_hosted_brokering_policy" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
    tozny_realm_broker_identity.broker_identity,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_broker_identity_credentials = tozny_realm_broker_identity.broker_identity.credentials
  use_tozny_hosted_broker = true
}

# A resource for provisioning a TozID identity to embed into another resource.
# Note this user specifies a password -- protect this sensitive information
# by ensuring state files are encrypted
resource "tozny_realm_identity" "machine_user_step" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name = tozny_realm.my_organizations_realm.realm_name
  username = "machine"
  password = "securePasswordFromSecretStore"
  email = "machine@exaple.com"
  broker_target_url = "http://localhost:8081/${tozny_realm.my_organizations_realm.realm_name}/recover"
}
```

## Argument Reference

### Top-Level Arguments

* `client_credentials_filepath` - (Optional) The filepath to Tozny client credentials for the Terraform provider to use when setting default groups. Omit if using `client_credentials_config`.
* `client_credentials_config` - (Optional) A JSON string containing Tozny client credentials for the provider to use when setting default groups. Omit if using `client_credentials_filepath`.
* `realm_name` - (Required) The name of the realm with which to associate the identity.
* `username - (Required) The username for this identity.
* `email - (Required) The email address associated with this identity.
* `client_registration_token - (Required) A registration token for the realm allowed to create identities.
* `broker_target_url - (Required) The base link for password resets.
* `password - (Required) The password for this identity. Ideally this comes from a secret store of some kind.
* `first_name` - (Optional) The first name associated with this identity.
* `last_name` - (Optional) The last name associated with this identity.
* `recovery_email_ttl` - (Optional) The length of time a recovery email is valid for.

## Attribute Reference
* `id` - The Tozny Client ID for the provisioned identity.