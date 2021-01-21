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
      version = ">=0.10.0"
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
# locals {
#   tozny_client_credentials_filepath = "./tozny_client_credentials.json"
# }

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
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
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
    tozny_account.autogenerated_tozny_account,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name = tozny_realm.my_organizations_realm.realm_name
  name = "broker${tozny_realm.my_organizations_realm.realm_name}"
  broker_identity_credentials_save_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
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
  realm_broker_identity_credentials_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
  use_tozny_hosted_broker = true
}

# A resource for creating an OIDC based realm application
resource "tozny_realm_application" "jenkins_oidc_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "jenkins-oid-app"
  name = "Jenkins"
  active = true
  protocol = "openid-connect"
  oidc_settings {
    allowed_origins = [ "https://jenkins.acme.com/allowed" ]
    access_type = "bearer-only"
    root_url = "https://jenkins.acme.com"
    standard_flow_enabled = true
    base_url = "https://jenkins.acme.com/baseurl"

  }
}

# A resource for retrieving the OIDC client secret for an application
# this resource is configured to bypass storing the secret in terraform
# but instead to a local file that can be secured and accessed outside of terraform
resource "tozny_realm_application_client_secret" "jenkins_oidc_client_secret_file_only" {
  depends_on = [
    tozny_realm.my_organizations_realm,
    tozny_realm_application.jenkins_oidc_application,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = tozny_realm_application.jenkins_oidc_application.application_id
  persist_client_secret_to_terraform = false
  client_secret_save_filepath = "./client_secret_file_only.txt"
}

# A resource for retrieving the OIDC client secret for an application
# this resource is configured to store the secret in terraform
# so it can be accessed by other terraform resources
resource "tozny_realm_application_client_secret" "jenkins_oidc_client_secret_terraform_only" {
  depends_on = [
    tozny_realm.my_organizations_realm,
    tozny_realm_application.jenkins_oidc_application,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = tozny_realm_application.jenkins_oidc_application.application_id
  persist_client_secret_to_terraform = true
}

# A resource for retrieving the OIDC client secret for an application
# this resource is configured to store the secret in terraform
# so it can be accessed by other terraform resources as well as
# to a local file that can be secured and accessed outside of terraform
resource "tozny_realm_application_client_secret" "jenkins_oidc_client_secret_terraform_and_file" {
  depends_on = [
    tozny_realm.my_organizations_realm,
    tozny_realm_application.jenkins_oidc_application,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = tozny_realm_application.jenkins_oidc_application.application_id
  persist_client_secret_to_terraform = true
  client_secret_save_filepath = "./client_secret_file_and_terraform.txt"
}
