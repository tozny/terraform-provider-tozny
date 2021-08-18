# Configure the Terraform runtime
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      # Pull signed provider binaries from
      # the Terraform hosted registry namespace
      # for Tozny registry.terraform.io/tozny
      source  = "tozny/tozny"
      # Pin Tozny provider version
      version = ">=0.12"
    }
  }
}

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
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
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
resource "tozny_realm_identity" "example_user" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name = tozny_realm.my_organizations_realm.realm_name
  username = "example"
  password = "securePasswordFromSecretStore"
  email = "luke+example@tozny.com"
  broker_target_url = "http://localhost:8081/${tozny_realm.my_organizations_realm.realm_name}/recover"
}

resource "tozny_realm_group" "default_group_1" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Default 1"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_group" "default_group_2" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Default 2"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_group" "non_default_group" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Not Default"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_group" "realm_admin_group" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Realm Administrators"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

# A data provider for fetching data about non-terraform managed applications (e.g. built-in applications)
data "tozny_realm_application" "realm_management_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "realm-management"
}

# A data provider for fetching data about non-terraform-managed realm application roles (e.g. built-in application roles)
data "tozny_realm_application_role" "realm_admin_role" {
  depends_on = [
    data.tozny_realm_application.realm_management_application
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = data.tozny_realm_application.realm_management_application.application_id
  name = "realm-admin"
}

# A data provider for fetching data about non-terraform-managed realm roles (e.g. built-in realm roles)
data "tozny_realm_role" "realm_offline_access_role" {
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  name = "offline_access"
}

# A resource mapping the realm-admin and offline access roles to the realm admin group
resource "tozny_realm_group_role_mappings" "admin_members_role_mappings" {
  depends_on = [
    tozny_realm_group.realm_admin_group,
    data.tozny_realm_application_role.realm_admin_role
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  group_id = tozny_realm_group.realm_admin_group.id
  application_role {
    application_id = data.tozny_realm_application.realm_management_application.application_id
    role_id = data.tozny_realm_application_role.realm_admin_role.application_role_id
    role_name = data.tozny_realm_application_role.realm_admin_role.name
  }
  realm_role {
    realm_id = tozny_realm.my_organizations_realm.id
    role_id = data.tozny_realm_role.realm_offline_access_role.realm_role_id
    role_name = data.tozny_realm_role.realm_offline_access_role.name
  }
}

resource "tozny_realm_default_groups" "default_groups" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  group_ids = [
    tozny_realm_group.default_group_1.group_id,
    tozny_realm_group.default_group_2.group_id,
  ]
}

resource "tozny_realm_identity_group_membership" "eg_membership" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  group_ids = concat(tozny_realm_default_groups.default_groups.group_ids, [tozny_realm_group.non_default_group.id, tozny_realm_group.realm_admin_group.id])
  identity_id = tozny_realm_identity.example_user.id
}
