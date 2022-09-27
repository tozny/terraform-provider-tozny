# Configure the Terraform runtime
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      # Pull signed provider binaries from
      # the Terraform hosted registry namespace
      # for Tozny registry.terraform.io/tozny
      source = "tozny/tozny"
      # Pin Tozny provider version
      version = ">=0.23.0"
    }
  }
}

# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint     = "http://platform.local.tozny.com:8000"
  account_username = "karthik@datacaliper.com"
  account_password = "Test@12345"
}

# A resource for provisioning the use of an external identity server
# for federating identities for a realm
resource "tozny_identity_provider" "azure_identity_provider" {
  realm_name                = "localtest"
  provider_id               = "oidc"
  display_name              = "Azure AD"
  enabled                   = true
  config {
    authorization_url       = "https://test-auth-url.com"
    token_url               = "https://test-auth-url.com"
    client_auth_method      = "client_secret_post"
    client_id               = "sdsdscscscdvdfdfdfdf"
    client_secret           = "asdasdsaxcdscdcddvdvfv"
  }
}
