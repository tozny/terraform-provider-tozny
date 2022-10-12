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

# A resource for mapping roles
resource "tozny_identity_provider_mapper" "idp_role_mapper" {
  realm_name                    = "localtest"
  alias                         = "azure-ad"  
  name                          = "Azure Role Map"
  identity_provider_mapper      = "oidc-role-idp-mapper"
  config {
        sync_mode   = "FORCE"
		    claim       = "roles"
		    claim_value = "Test.Role"
		    role        = "FirstRole"
  }
}
