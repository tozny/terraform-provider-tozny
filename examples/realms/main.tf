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
      version = ">=0.9.0"
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
