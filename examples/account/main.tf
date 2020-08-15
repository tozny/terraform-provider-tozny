# Specify provider configuration for Terraform
# directing it to pull the Tozny provider from somewhere other
# than the default registry
# Not needed now that Hashicorp approved us for publishing to the hashicorp
# registry under the tozny namespace
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      source  = "terraform.tozny.com/tozny/tozny"
      version = ">=0.0.1"
    }
  }
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_username_salt" {
  length = 8
  special = false
}

# Include the Tozny Terraform provider for this module
provider "tozny" {
  api_endpoint = "http://platform.local.tozny.com:8000"
  account_username = "account+${random_string.account_username_salt.result}@tozny.terraform.com"
  account_password = "readymyenvironment"
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_email_salt" {
  length = 8
  special = false
}

# # A resource for provisioning a Tozny account from explicit configuration
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

# A resource for provisioning a Tozny account from explicit configuration
# read from a local file.
# Tied for 1st most secure method in that no private information has to be passed to terraform
resource "tozny_account" "pregenerated_tozny_account_credentials_from_file" {
  autogenerate_account_credentials = false
  account_credentials_filepath = "account.json"
  client_credentials_save_filepath = "./tozny_client_credentials.json"
}

