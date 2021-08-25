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
      version = ">=0.12.1"
    }
  }
}

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

# A resource for provisioning the use of an external ldap server
# for providing and federating identities for a realm
resource "tozny_realm_provider" "ldap_identity_provider" {
 depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  provider_type = "ldap"
  name = "LDAP Identity Provider"
  active = true
  import_identities = true
  priority = 0
  connection_settings {
    type = "ad"
    identity_name_attribute = "cn"
    rdn_attribute = "cn"
    uuid_attribute = "objectGUID"
    identity_object_classes = ["person", "organizationalPerson", "user"]
    connection_url = "ldap://test.local"
    identity_dn = "cn=users,dc=tozny,dc=local"
    authentication_type = "simple"
    bind_dn = "TOZNY\\administrator"
    bind_credential = "password"
    search_scope = 1
    trust_store_spi_mode = "ldapsOnly"
    connection_pooling = true
    pagination = true
  }
}
