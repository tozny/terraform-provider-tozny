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
      version = ">=0.14.0"
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
    implicit_flow_enabled = true
    direct_access_grants_enabled = true
    base_url = "https://jenkins.acme.com/baseurl"
  }
}

# A resource for creating a SAML based realm application
resource "tozny_realm_application" "aws_saml_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "aws-saml-app"
  name = "AWS"
  active = true
  protocol = "saml"
  saml_settings {
    allowed_origins = [ "https://example.frontend.com" ]
    default_endpoint = "https://samuel/saml/iam"
    include_authn_statement = true
    include_one_time_use_condition = true
    sign_documents = true
    sign_assertions = true
    client_signature_required = true
    force_post_binding = true
    force_name_id_format = true
    name_id_format = "name_id_format"
    idp_initiated_sso_url_name = "sso_url_name"
    assertion_consumer_service_post_binding_url = "post_binding_url"
  }
}

# A data provider for fetching information about non-terraform managed applications (e.g. built-in applications)
data "tozny_realm_application" "account_application" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "account"
}
