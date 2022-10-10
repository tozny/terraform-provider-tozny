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
  display_name              = "Azure AD"
  alias                     = "azure-ad-1"
  enabled                   = true
  config {
    authorization_url       = "https://test-eidp.com/auth"
    token_url               = "https://test-eidp.com/token"
    client_auth_method      = "client_secret_post"
    client_id               = "sdsdscscscdvdfdfdfdf"
    client_secret           = "asdasdsaxcdscdcddvdvfv"
    default_scope           = "email profile openid"
  }
}

resource "tozny_identity_provider_mapper" "idp_role_mapper" {
  depends_on = [
    tozny_identity_provider.azure_identity_provider,
  ]
  realm_name                    = "localtest"
  alias                         = "azure-ad-1"  
  name                          = "Azure Role Map"
  identity_provider_mapper      = "oidc-role-idp-mapper"
  config {
        sync_mode   = "FORCE"
		    claim       = "roles"
		    claim_value = "Test.Role"
		    role        = "FirstRole"
  }
}

resource "tozny_identity_provider" "okta_identity_provider" {
  realm_name                = "localtest"
  display_name              = "Okta"
  alias                     = "okta"
  enabled                   = true
  config {
    authorization_url       = "https://dev-49909711.okta.com/oauth2/default/v1/authorize"
    token_url               = "https://dev-49909711.okta.com/oauth2/default/v1/token"
    client_auth_method      = "client_secret_post"
    client_id               = "0oa66z9mi3F5akVdh5d7"
    client_secret           = "NnSR7HrXL2TQx5r3pwmiBBZ6qr6nwL5lMUmt8iIJ"
    default_scope           = "email profile openid"
  }
}