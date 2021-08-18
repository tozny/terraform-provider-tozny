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
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = random_string.random_realm_name.result
  sovereign_name = "Administrator"
}

resource "tozny_realm_group" "default_group_1" {
  depends_on = [
    tozny_realm.my_organizations_realm
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Default 1"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_group" "default_group_2" {
  depends_on = [
    tozny_realm.my_organizations_realm
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Default 2"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_group" "non_default_group" {
  depends_on = [
    tozny_realm.my_organizations_realm
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Not Default"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

resource "tozny_realm_default_groups" "default_groups" {
  depends_on = [
    tozny_realm_group.default_group_1,
    tozny_realm_group.default_group_2
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  group_ids = [
    tozny_realm_group.default_group_1.group_id,
    tozny_realm_group.default_group_2.group_id,
  ]
}
