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
      version = ">=0.27.0"
    }
  }
}

# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint     = "http://platform.local.tozny.com:8000"
  account_username = "test-emails-group+${random_string.account_username_salt.result}@tozny.com"
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_username_salt" {
  length  = 8
  special = false
}

# Generate a random string for use in creating
# realms across environments or executions conflict free
resource "random_string" "random_realm_name" {
  length  = 8
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
  persist_credentials_to           = "file"
  client_credentials_save_filepath = local.tozny_client_credentials_filepath
}

# A resource for provisioning a TozID Realm
# using local file based credentials for a Tozny Client with permissions to manage Realms.
resource "tozny_realm" "my_organizations_realm" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = random_string.random_realm_name.result
  sovereign_name              = "Administrator"
}

# A resource for provisioning a token that can be used to register Tozny clients
# such as a realm broker identity
resource "tozny_client_registration_token" "realm_registration_token" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
  client_credentials_filepath       = local.tozny_client_credentials_filepath
  name                              = "${tozny_realm.my_organizations_realm.realm_name}DefaultClientRegistrationToken"
  allowed_registration_client_types = ["general", "identity", "broker"]
  enabled                           = true
  one_time_use                      = false
}

# A resource for provisioning an identity that can be used to delegate authority for
# realm activities such as email recovery
resource "tozny_realm_broker_identity" "broker_identity" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
  ]
  client_credentials_filepath               = local.tozny_client_credentials_filepath
  client_registration_token                 = tozny_client_registration_token.realm_registration_token.token
  realm_name                                = tozny_realm.my_organizations_realm.realm_name
  name                                      = "broker${tozny_realm.my_organizations_realm.realm_name}"
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
  client_credentials_filepath                = local.tozny_client_credentials_filepath
  realm_broker_identity_credentials_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
  use_tozny_hosted_broker                    = true
}

# A resource for creating an OIDC based realm application
resource "tozny_realm_application" "jenkins_oidc_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  client_id                   = "jenkins-oid-app"
  name                        = "Jenkins"
  active                      = true
  protocol                    = "openid-connect"
  oidc_settings {
    allowed_origins = ["https://example.frontend.com"]
    root_url        = "https://jenkins.acme.com"
  }
}

# A resource for creating a SAML based realm application
resource "tozny_realm_application" "aws_saml_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  client_id                   = "aws-saml-app"
  name                        = "AWS"
  active                      = true
  protocol                    = "saml"
  saml_settings {
    allowed_origins                             = ["https://example.frontend.com"]
    default_endpoint                            = "https://samuel/saml/iam"
    include_authn_statement                     = true
    include_one_time_use_condition              = true
    sign_documents                              = true
    sign_assertions                             = true
    client_signature_required                   = true
    force_post_binding                          = true
    force_name_id_format                        = true
    name_id_format                              = "name_id_format"
    idp_initiated_sso_url_name                  = "sso_url_name"
    assertion_consumer_service_post_binding_url = "post_binding_url"
  }
}

# A resource for creating an application oidc mapper for identity policy attribute
resource "tozny_realm_application_mapper" "oidc_client_policy_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.jenkins_oidc_application.application_id
  name                        = "Client Policy"
  protocol                    = "openid-connect"
  mapper_type                 = "oidc-user-attribute-mapper"
  add_to_user_info            = true
  add_to_id_token             = true
  add_to_access_token         = true
  multivalued                 = false
  aggregate_attribute_values  = false
  user_attribute              = "policy"
  claim_json_type             = "String"
  token_claim_name            = "policy"
}

# A resource for creating an application oidc mapper for identity group membership
resource "tozny_realm_application_mapper" "oidc_group_memebership_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.jenkins_oidc_application.application_id
  name                        = "group-membership"
  protocol                    = "openid-connect"
  mapper_type                 = "oidc-group-membership-mapper"
  add_to_user_info            = true
  add_to_id_token             = true
  add_to_access_token         = true
  claim_json_type             = "String"
  token_claim_name            = "group-membership"
  full_group_path             = true
}

# A resource for creating an application saml mapper for identity name attribute
resource "tozny_realm_application_mapper" "saml_name_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.aws_saml_application.application_id
  name                        = "name"
  protocol                    = "saml"
  mapper_type                 = "saml-user-property-mapper"
  property                    = "Username"
  friendly_name               = "Username"
  saml_attribute_name         = "name"
  saml_attribute_name_format  = "Basic"
}

# A resource for creating an application saml mapper for identity roles attribute
resource "tozny_realm_application_mapper" "saml_roles_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.aws_saml_application.application_id
  name                        = "roles"
  protocol                    = "saml"
  mapper_type                 = "saml-role-list-mapper"
  role_attribute_name         = "roles"
  friendly_name               = "Roles"
  saml_attribute_name_format  = "Basic"
  single_role_attribute       = true
}
# A resource for creating an application oidc mapper for identity policy realm roles mapper
resource "tozny_realm_application_mapper" "realm_roles_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.jenkins_oidc_application.application_id
  name                        = "Realm Role Policy Mapper"
  protocol                    = "openid-connect"
  mapper_type                 = "oidc-usermodel-realm-role-mapper"
  add_to_user_info            = true
  add_to_id_token             = true
  add_to_access_token         = true
  multivalued                 = false
  claim_json_type             = "String"
  token_claim_name            = "policy"
  realm_role_prefix           = "String"
}

# A resource for creating an application oidc mapper for identity policy client roles mapper
resource "tozny_realm_application_mapper" "client_roles_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.jenkins_oidc_application.application_id
  name                        = "Client Role Policy Mapper"
  protocol                    = "openid-connect"
  mapper_type                 = "oidc-usermodel-client-role-mapper"
  add_to_user_info            = true
  add_to_id_token             = true
  add_to_access_token         = true
  multivalued                 = false
  claim_json_type             = "String"
  token_claim_name            = "policy"
  client_role_prefix          = "string"
  client_id                   = "tozid-realm-idp"
}

# A resource for creating an application oidc mapper for identity policy user attribute
resource "tozny_realm_application_mapper" "user_attributes_mapper" {
  depends_on = [
    tozny_realm_application.aws_saml_application
  ]
  client_credentials_filepath = local.tozny_client_credentials_filepath
  realm_name                  = tozny_realm.my_organizations_realm.realm_name
  application_id              = tozny_realm_application.jenkins_oidc_application.application_id
  name                        = "User Attributes Policy Mapper"
  protocol                    = "openid-connect"
  mapper_type                 = "oidc-usermodel-attribute-mapper"
  add_to_user_info            = true
  add_to_id_token             = true
  add_to_access_token         = true
  multivalued                 = false
  aggregate_attribute_values  = false
  user_attribute              = "policy"
  claim_json_type             = "String"
  token_claim_name            = "policy"
}
