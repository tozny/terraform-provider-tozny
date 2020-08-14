# Specify provider configuration for Terraform
# directing it to pull the Tozny provider from somewhere other
# than the default registry
# Won't be needed once Hashicorp approves us for publishing or we
# host our own registry @terraform.tozny.com
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      source  = "terraform.tozny.com/tozny/tozny"
      version = ">=0.0.1"
    }
  }
}
# Include the Tozny Terraform provider for this module
provider "tozny" {

}

# A resource for creating a Tozny account
resource "tozny_account" "my_tozny_account" {
  profile {
    id = 1
  }
  account {
    id = 2
  }
}
